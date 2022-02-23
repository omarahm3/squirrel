package main

import (
	"log"

	"github.com/gorilla/websocket"
)

const WEBSOCKET_URL string = "ws://localhost:3000/ws"

type LogMessage struct {
	Line string `json:"line"`
}

type Message struct {
	Id      string      `json:"id"`
	Local   bool        `json:"local"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
}

func InitClient(input chan string) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(WEBSOCKET_URL, nil)

	if err != nil {
		log.Fatal("Error connecting to websocket server:", err)
	}

	return connection
}

// Needed to receive server events
// Right now we do nothing, but its here to avoid errors on the protocol
func HandleIncomingMessages(connection *websocket.Conn) {
	defer func() {
    connection.Close()
	}()

	for {
		_, _, err := connection.ReadMessage()

		if err != nil {
			break
		}
	}
}

func HandleWebsocketClose(connection *websocket.Conn) {
	err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	if err != nil {
		log.Fatal("Error during closing websocket:", err)
		return
	}
}
