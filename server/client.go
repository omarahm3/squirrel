package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	id         string
	connection *websocket.Conn
	hub        *Hub
	send       chan []byte
}

func (client *Client) ReadPump() {
	defer func() {
		client.hub.unregister <- client
		client.connection.Close()
	}()

	for {
		var message LogMessage
		err := client.connection.ReadJSON(&message)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected server close: ", err)
			}

			break
		}

		log.Println(message)

		err = client.connection.WriteJSON(message)

		if err != nil {
			break
		}
	}
}
