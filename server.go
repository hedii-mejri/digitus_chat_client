package main

import (
	"./endpoint/json"
	"./util"
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
)

const LOBBY = "lobby"

// program main
func main() {
	// start the chat server
	properties := util.LoadConfig()
	psock, err := net.Listen("tcp", ":"+properties.Port)
	util.CheckForError(err, "Can't create server")

	fmt.Printf("Chat server started on port %v...\n", properties.Port)
	fmt.Printf("To create a client, provide the USERNAME\n")

	// start the JSON endpoing server
	go json.Start()

	for {
		// accept connections
		conn, err := psock.Accept()
		util.CheckForError(err, "Can't accept connections")

		// keep track of the client details
		client := util.Client{Connection: conn, Room: LOBBY, Properties: properties}
		client.Register()

		channel := make(chan string)
		go waitForInput(channel, &client)
		go handleInput(channel, &client, properties)

		util.SendClientMessage("ready", properties.Port, &client, true, properties)
	}
}

func waitForInput(out chan string, client *util.Client) {
	defer close(out)

	reader := bufio.NewReader(client.Connection)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			// Remove the client
			client.Close(true)
			return
		}
		out <- string(line)
	}
}

func handleInput(in <-chan string, client *util.Client, props util.Properties) {

	for {
		message := <-in
		if message != "" {
			message = strings.TrimSpace(message)
			action, body := getAction(message)

			if action != "" {
				switch action {

				// the user has submitted a message
				case "message":
					util.SendClientMessage("message", body, client, false, props)

				// the user has provided their username (initialization handshake)
				case "user":
					client.Username = body
					util.SendClientMessage("connect", "", client, false, props)

				// the user is disconnecting
				case "disconnect":
					client.Close(false)

				// the user is disconnecting
				case "ignore":
					client.Ignore(body)
					util.SendClientMessage("ignoring", body, client, false, props)

				// the user is entering a room
				case "enter":
					if body != "" {
						client.Room = body
						util.SendClientMessage("enter", body, client, false, props)
					}

				// the user is leaving the current room
				case "leave":
					if client.Room != LOBBY {
						util.SendClientMessage("leave", client.Room, client, false, props)
						client.Room = LOBBY
					}

				default:
					util.SendClientMessage("unrecognized", action, client, true, props)
				}
			}
		}
	}
}

// parse data and return values
func getAction(message string) (string, string) {
	actionRegex, _ := regexp.Compile(`^\/([^\s]*)\s*(.*)$`)
	res := actionRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		return res[0][1], res[0][2]
	}
	return "", ""
}
