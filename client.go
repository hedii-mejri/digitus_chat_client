package main

import (
	"./util"
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

var standardInputMessageRegex, _ = regexp.Compile(`^\/([^\s]*)\s*(.*)$`)
var chatServerResponseRegex, _ = regexp.Compile(`^\/([^\s]*)\s?(?:\[([^\]]*)\])?\s*(.*)$`)

type Command struct {
	Command, Username, Body string
}

// Main container
func main() {
	username, properties := getConfig()

	conn, err := net.Dial("tcp", properties.Hostname+":"+properties.Port)
	util.CheckForError(err, "Connection refused")
	defer conn.Close()

	go watchForConnectionInput(username, properties, conn)
	for true {
		watchForConsoleInput(conn)
	}
}

func getConfig() (string, util.Properties) {
	if len(os.Args) >= 2 {
		username := os.Args[1]
		properties := util.LoadConfig()
		return username, properties
	} else {
		println("Please the username as the first parameter ")
		os.Exit(1)
		return "", util.Properties{}
	}
}
func watchForConsoleInput(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for true {
		message, err := reader.ReadString('\n')
		util.CheckForError(err, "Lost console connection")

		message = strings.TrimSpace(message)
		if message != "" {
			command := parseInput(message)

			if command.Command == "" {
				sendCommand("message", message, conn)
			} else {
				switch command.Command {

				// enter a room
				case "enter":
					sendCommand("enter", command.Body, conn)

				// ignore someone
				case "ignore":
					sendCommand("ignore", command.Body, conn)

				// leave a room
				case "leave":
					sendCommand("leave", "", conn)

				// disconnect from the chat server
				case "disconnect":
					sendCommand("disconnect", "", conn)

				default:
					fmt.Printf("Unknown command \"%s\"\n", command.Command)
				}
			}
		}
	}
}

func watchForConnectionInput(username string, properties util.Properties, conn net.Conn) {
	reader := bufio.NewReader(conn)

	for true {
		message, err := reader.ReadString('\n')
		util.CheckForError(err, "Lost server connection")
		message = strings.TrimSpace(message)
		if message != "" {
			Command := parseCommand(message)
			switch Command.Command {

			// send out the username
			case "ready":
				sendCommand("user", username, conn)

			// the user has connected to the chat server
			case "connect":
				fmt.Printf(properties.HasEnteredTheLobbyMessage+"\n", Command.Username)

			// the user has disconnected
			case "disconnect":
				fmt.Printf(properties.HasLeftTheLobbyMessage+"\n", Command.Username)

			// the user has entered a room
			case "enter":
				fmt.Printf(properties.HasEnteredTheRoomMessage+"\n", Command.Username, Command.Body)

			// the user has left a room
			case "leave":
				fmt.Printf(properties.HasLeftTheRoomMessage+"\n", Command.Username, Command.Body)

			// the user has sent a message
			case "message":
				if Command.Username != username {
					fmt.Printf(properties.ReceivedAMessage+"\n", Command.Username, Command.Body)
				}

			// the user has connected to the chat server
			case "ignoring":
				fmt.Printf(properties.IgnoringMessage+"\n", Command.Body)
			}
		}
	}
}

// In the form of /command {body content}
func sendCommand(command string, body string, conn net.Conn) {
	message := fmt.Sprintf("/%v %v\n", util.Encode(command), util.Encode(body))
	conn.Write([]byte(message))
}

func parseInput(message string) Command {
	res := standardInputMessageRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		// there is a command
		return Command{
			Command: res[0][1],
			Body:    res[0][2],
		}
	} else {
		return Command{
			Body: util.Decode(message),
		}
	}
}

// look for "/Command [name] body contents" (Optionnal)
func parseCommand(message string) Command {
	res := chatServerResponseRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		// we've got a match
		return Command{
			Command:  util.Decode(res[0][1]),
			Username: util.Decode(res[0][2]),
			Body:     util.Decode(res[0][3]),
		}
	} else {
		return Command{}
	}
}
