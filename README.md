# digitus chat client
Simple chat client and server for DigitUs


Chat Server
-----------

Clone the repo
```
> git clone https://github.com/jhudson8/golang-chat-example.git
> cd digitus_chat_client-master
```

Edit the server/client configuration as you need (```config.json```)
```
{
  "Port": "5555",
  "JSONEndpointPort": "8080",
  "Hostname": "localhost",
  "HasEnteredTheRoomMessage": "[%s] has entered the room \"%s\"",
  "HasLeftTheRoomMessage": "[%s] has left the room \"%s\"",
  "HasEnteredTheLobbyMessage": "[%s] has entered the lobby",
  "HasLeftTheLobbyMessage": "[%s] has left the lobby",
  "IgnoringMessage": "You are ignoring %s",
  "ReceivedAMessage": "[%s] says: %s",
  "LogFile": ""
}

```

Start the server
```
> go run server.go
```


Chat Client
-----------
In other terminal windows, create clients
```
> go run client.go {username}
```

You can send commands or messages.  Commands begin with ```/```
The commands are available

* ```/enter SomeRoom```: enter a private room (only messages from others in the same private room will be visible) and this automatically creates the room if unavailable.
* ```leave```: leave a private room to go back to the main lobby ```/leave```
* ```ignore```: ignore another user ```/ignore SomeUser```
* ```disconnect```: disconnect from the chat server

JSON Endpoint
----------
The JSON endpoint port can be configured using the ```JSONEndpointPort``` port (by default, 8080).  When the chat server is stated, the following endpoints are available

* ```/messages/all```: all messages
* ```/messages/search/{search term}```: example ```localhost:8080/messages/search/hello```
* ```/messages/user/{username}```: example ```localhost:8080/messages/user/joe```

PS: Previously logged messages will not be evaluated; only current session messages will be handled.


Chat Log
----------
Log files are in CSV format with the columns shown below.  You *must* set the ```LogFile``` config value to be the absolute file location or no logs will be created.

1. ***username***: the user that performed the action
2. ***action***: the action that was taken (```message```/```enter```/```leave```/```ignore```/```connect```/```disconnect```)
3. ***value***: the chat message or room that was entered or left
4. ***timestamp***: example ```Mar 12 2017 09.13.05 +0100 EDT```
5. ***ip***: example ```127.0.0.1:53594```
