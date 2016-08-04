package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

// time format for log files and JSON response
const TIME_LAYOUT = "Aug 1 2016 15.04.05 +0100 MST"

// Encoding Data to Clients
var ENCODING_UNENCODED_TOKENS = []string{"%", ":", "[", "]", ",", "\""}
var ENCODING_ENCODED_TOKENS = []string{"%25", "%3A", "%5B", "%5D", "%2C", "%22"}
var DECODING_UNENCODED_TOKENS = []string{":", "[", "]", ",", "\"", "%"}
var DECODING_ENCODED_TOKENS = []string{"%3A", "%5B", "%5D", "%2C", "%22", "%25"}

// User Information
type Client struct {
	// the client's connection
	Connection net.Conn
	Username   string
	Room       string
	ignoring   []string
	// Configuration File
	Properties Properties
}

func (client *Client) Close(doSendMessage bool) {
	if doSendMessage {
		SendClientMessage("disconnect", "", client, false, client.Properties)
	}
	client.Connection.Close()
	clients = removeEntry(client, clients)
}

func (client *Client) Register() {
	clients = append(clients, client)
}

func (client *Client) Ignore(username string) {
	client.ignoring = append(client.ignoring, username)
}

func (client *Client) IsIgnoring(username string) bool {
	for _, value := range client.ignoring {
		if value == username {
			return true
		}
	}
	return false
}

// lLogging Container
type Action struct {
	// "message", "leave", "enter", "connect", "disconnect"
	Command   string `json:"command"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
}

type Properties struct {
	Hostname                  string
	Port                      string
	JSONEndpointPort          string
	HasEnteredTheRoomMessage  string
	HasLeftTheRoomMessage     string
	HasEnteredTheLobbyMessage string
	HasLeftTheLobbyMessage    string
	ReceivedAMessage          string
	IgnoringMessage           string
	LogFile                   string
}

var actions = []Action{}

var config = Properties{}

var clients []*Client

func LoadConfig() Properties {
	if config.Port != "" {
		return config
	}
	pwd, _ := os.Getwd()

	payload, err := ioutil.ReadFile(pwd + "/config.json")
	CheckForError(err, "Cannot read config file")

	var dat map[string]interface{}
	err = json.Unmarshal(payload, &dat)
	CheckForError(err, "Invalid JSON in config file")

	var rtn = Properties{
		Hostname:                  dat["Hostname"].(string),
		Port:                      dat["Port"].(string),
		JSONEndpointPort:          dat["JSONEndpointPort"].(string),
		HasEnteredTheRoomMessage:  dat["HasEnteredTheRoomMessage"].(string),
		HasLeftTheRoomMessage:     dat["HasLeftTheRoomMessage"].(string),
		HasEnteredTheLobbyMessage: dat["HasEnteredTheLobbyMessage"].(string),
		HasLeftTheLobbyMessage:    dat["HasLeftTheLobbyMessage"].(string),
		ReceivedAMessage:          dat["ReceivedAMessage"].(string),
		IgnoringMessage:           dat["IgnoringMessage"].(string),
		LogFile:                   dat["LogFile"].(string),
	}
	config = rtn
	return rtn
}

func removeEntry(client *Client, arr []*Client) []*Client {
	rtn := arr
	index := -1
	for i, value := range arr {
		if value == client {
			index = i
			break
		}
	}

	if index >= 0 {
		rtn = make([]*Client, len(arr)-1)
		copy(rtn, arr[:index])
		copy(rtn[index:], arr[index+1:])
	}

	return rtn
}

func SendClientMessage(messageType string, message string, client *Client, thisClientOnly bool, props Properties) {

	if thisClientOnly {
		message = fmt.Sprintf("/%v", messageType)
		fmt.Fprintln(client.Connection, message)

	} else if client.Username != "" {
		LogAction(messageType, message, client, props)

		// construct the payload to be sent to clients
		payload := fmt.Sprintf("/%v [%v] %v", messageType, client.Username, message)

		for _, _client := range clients {
			// Write the message to the client
			if (thisClientOnly && _client.Username == client.Username) ||
				(!thisClientOnly && _client.Username != "") {

				if messageType == "message" && client.Room != _client.Room || _client.IsIgnoring(client.Username) {
					continue
				}

				fmt.Fprintln(_client.Connection, payload)
			}
		}
	}
}

func CheckForError(err error, message string) {
	if err != nil {
		println(message+": ", err.Error())
		os.Exit(1)
	}
}

// double quote the single quotes
func EncodeCSV(value string) string {
	return strings.Replace(value, "\"", "\"\"", -1)
}

// Encoding data over HTTP
func Encode(value string) string {
	return replace(ENCODING_UNENCODED_TOKENS, ENCODING_ENCODED_TOKENS, value)
}

// Decoding data over HTTP
func Decode(value string) string {
	return replace(DECODING_ENCODED_TOKENS, DECODING_UNENCODED_TOKENS, value)
}

func replace(fromTokens []string, toTokens []string, value string) string {
	for i := 0; i < len(fromTokens); i++ {
		value = strings.Replace(value, fromTokens[i], toTokens[i], -1)
	}
	return value
}

func LogAction(action string, message string, client *Client, props Properties) {
	ip := client.Connection.RemoteAddr().String()
	timestamp := time.Now().Format(TIME_LAYOUT)

	actions = append(actions, Action{
		Command:   action,
		Content:   message,
		Username:  client.Username,
		IP:        ip,
		Timestamp: timestamp,
	})

	if props.LogFile != "" {
		if message == "" {
			message = "N/A"
		}
		fmt.Printf("logging values %s, %s, %s\n", action, message, client.Username)

		logMessage := fmt.Sprintf("\"%s\", \"%s\", \"%s\", \"%s\", \"%s\"\n",
			EncodeCSV(client.Username), EncodeCSV(action), EncodeCSV(message),
			EncodeCSV(timestamp), EncodeCSV(ip))

		f, err := os.OpenFile(props.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			err = ioutil.WriteFile(props.LogFile, []byte{}, 0600)
			f, err = os.OpenFile(props.LogFile, os.O_APPEND|os.O_WRONLY, 0600)
			CheckForError(err, "Cannot create log file")
		}

		defer f.Close()
		_, err = f.WriteString(logMessage)
		CheckForError(err, "Cannot write to log file")
	}
}

func QueryMessages(actionType string, search string, username string) []Action {

	isMatch := func(action Action) bool {
		if actionType != "" && action.Command != actionType {
			return false
		}
		if search != "" && !strings.Contains(action.Content, search) {
			return false
		}
		if username != "" && action.Username != username {
			return false
		}
		return true
	}

	rtn := make([]Action, 0, len(actions))

	for _, value := range actions {
		if isMatch(value) {
			rtn = append(rtn, value)
		}
	}

	return rtn
}
