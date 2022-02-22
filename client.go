package main

import (
	"log"

	"github.com/gorilla/websocket"
)

const WEBSOCKET_URL string = "ws://localhost:3000/ws"

func InitClient(input chan string) *websocket.Conn {
	connection, _, err := websocket.DefaultDialer.Dial(WEBSOCKET_URL, nil)

	if err != nil {
		log.Fatal("Error connecting to websocket server:", err)
	}

	return connection
}

func SendMessage(connection *websocket.Conn, message string) {
	err := connection.WriteMessage(websocket.TextMessage, []byte(message))

	if err != nil {
		log.Println("Error during sending message to websocket:", err)
		return
	}
}

func HandleWebsocketClose(connection *websocket.Conn) {
	err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	if err != nil {
		log.Fatal("Error during closing websocket:", err)
		return
	}
}
