package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT       = 10 * time.Second
	PONG_WAIT        = 60 * time.Second
	PING_PERIOD      = (PONG_WAIT * 9) / 10
	MAX_MESSAGE_SIZE = 1024
)

var (
	newLine = []byte{'\n'}
	// space   = []byte{' '}
)

type Client struct {
	id         string
	local      bool
	connection *websocket.Conn
	hub        *Hub
	send       chan []byte
	peerId     string
	active     bool
}

func (client *Client) ReadPump() {
	defer func() {
		log.Println("ReadPump::: Removing client, and closing connection")
		client.hub.unregister <- client
		client.connection.Close()
	}()

	client.connection.SetReadLimit(MAX_MESSAGE_SIZE)
	client.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
	client.connection.SetPongHandler(func(appData string) error {
		client.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})

	for {
		message, err := HandleMessage(client)

		if err != nil {
			break
		}

		// If this is a local client sending us the very first request
		if !client.local && message.Id != "" {
			oldId := client.id
			client.id = message.Id

			client.hub.update <- struct {
				id     string
				client *Client
			}{oldId, client}
		}
	}
}

func (client *Client) WritePump() {
	ticker := time.NewTicker(PING_PERIOD)

	defer func() {
		ticker.Stop()
		client.connection.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err != nil {
				return
			}

			if !ok {
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := client.connection.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			_, err = writer.Write(message)

			if err != nil {
				return
			}

			// Handle queued messages

			for i := 0; i < len(client.send); i++ {
				_, err = writer.Write(newLine)

				if err != nil {
					return
				}

				_, err = writer.Write(<-client.send)

				if err != nil {
					return
				}
			}

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			log.Println("Sending Ping Message")
			err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err != nil {
				log.Fatal("Error setting write deadline", err)
				return
			}

			if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Fatal("Error sending Ping message", err)
				return
			}
		}
	}
}

func HandleNewLogLine(client *Client, message LogMessage) {
	client.hub.broadcast <- struct {
		message  []byte
		clientId string
	}{
		message:  []byte(message.Line),
		clientId: client.id,
	}
}

func HandleIdentityMessage(client *Client, message Message, payload IdentityMessage) {
	var updateId string

	// In case this is a local peer
	if payload.Local {
		updateId = client.id
		client.id = message.Id
		client.local = payload.Local
		client.peerId = ""
	} else {
		if payload.PeerId == "" {
			log.Println("PeerId must not be empty, discarding")
			return
		}

		updateId = client.id
		client.peerId = payload.PeerId
	}

	client.active = true

	client.hub.update <- struct {
		id     string
		client *Client
	}{updateId, client}

	log.Println("--------------------------")
	log.Printf("ID: %s", client.id)
	log.Printf("Peer ID: %s", client.peerId)
	log.Printf("Active: %t", client.active)
	log.Printf("Local: %t", client.local)
	log.Println("--------------------------")
}

func HandleMessage(client *Client) (Message, error) {
	var message Message
	err := client.connection.ReadJSON(&message)

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Println("Unexpected server close: ", err)
		}

		return Message{}, err
	}

	switch message.Event {
	case "identity":
		identityMessage := IdentityMessage{}
		data, err := json.Marshal(message.Payload)

		if err != nil {
			log.Fatalln("HandleMessage.Identity::: Unexpected error (marshal):", err)
			return Message{}, err
		}

		err = json.Unmarshal([]byte(data), &identityMessage)

		if err != nil {
			log.Fatalln("HandleMessage.Identity::: Unexpected error (unmarshal):", err)
			return Message{}, err
		}

		HandleIdentityMessage(client, message, identityMessage)

	case "log_line":
		if !client.active {
			log.Println("HandleMessage.LogLine::: Unknown client, ignoring messages")
			return Message{}, errors.New("Client is unknown, ignoring messages")
		}

		logMessage := LogMessage{}
		data, err := json.Marshal(message.Payload)

		if err != nil {
			log.Fatalln("HandleMessage.LogLine::: Unexpected error (marshal):", err)
			return Message{}, err
		}

		err = json.Unmarshal([]byte(data), &logMessage)

		if err != nil {
			log.Fatalln("HandleMessage.LogLine::: Unexpected error (unmarshal):", err)
			return Message{}, err
		}

		HandleNewLogLine(client, logMessage)
	}

	return message, nil
}
