package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	id         string
	local      bool
	connection *websocket.Conn
	hub        *Hub
	send       chan []byte
}

type Message struct {
	Id      string      `json:"id"`
	Local   bool        `json:"local"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
}

type LogMessage struct {
	Line string `json:"line"`
}

func (client *Client) ReadPump() {
	defer func() {
		client.hub.unregister <- client
		client.connection.Close()
	}()

	for {
		message, err := HandleMessage(client)

		if err != nil {
			break
		}

		// If this is a local client sending us the very first request
		if !client.local && message.Id != "" {
			client.id = message.Id
			client.local = message.Local
		}

		err = client.connection.WriteJSON(message)

		if err != nil {
			break
		}
	}
}

func HandleNewLogLine(client *Client, message LogMessage) {
	log.Println(message)
}

func HandleMessage(client *Client) (Message, error) {
	var message Message
	err := client.connection.ReadJSON(&message)

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("Unexpected server close: ", err)
		}
	}

	switch message.Event {
	case "log_line":
		logMessage := LogMessage{}
		data, err := json.Marshal(message.Payload)

		if err != nil {
			log.Fatalln("#1# Unexpected error occurred:", err)
			return Message{}, err
		}

		err = json.Unmarshal([]byte(data), &logMessage)

		if err != nil {
			log.Fatalln("#2# Unexpected error occurred:", err)
			return Message{}, err
		}

		HandleNewLogLine(client, logMessage)
	}

	return message, nil
}
