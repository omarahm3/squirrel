package server

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	WRITE_WAIT     = 10 * time.Second
	PONG_WAIT      = 60 * time.Second
	PING_PERIOD    = (PONG_WAIT * 9) / 10
	EVENT_IDENTITY = "identity"
	EVENT_LOG_LINE = "log_line"
)

var (
	newLine = []byte{'\n'}
)

type Client struct {
	id          string
	broadcaster bool
	subscriber  bool
	connection  *websocket.Conn
	hub         *Hub
	send        chan []byte
	peerId      string
	active      bool
}

func (client *Client) ReadPump() {
	defer func() {
		zap.S().Info("Removing client")
		client.hub.unregister <- client
		zap.S().Info("Closing client connection")
		client.connection.Close()
	}()

	readDeadline := time.Now().Add(PONG_WAIT)

	zap.S().Infow(
		"Configuring client websocket connection",
		"readLimit", options.MaxMessageSize,
		"readDeadLine", readDeadline,
		"pongWait", PONG_WAIT,
	)

	client.connection.SetReadLimit(options.MaxMessageSize)
	client.connection.SetReadDeadline(readDeadline)
	client.connection.SetPongHandler(func(_ string) error {
		client.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})

	for {
		message, err := HandleMessage(client)

		if err != nil {
			zap.L().Error("Error handling message, breaking connection loop", zap.Error(err))
			break
		}

		// If this is a broadcaster client sending us the very first request
		if !client.broadcaster && message.Id != "" {
			zap.S().Debugw(
				"Broadcaster client connection, updating client ID",
				"oldId", client.id,
				"newId", message.Id,
			)

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
		zap.S().Info("Stopping ticker")
		ticker.Stop()
		zap.S().Info("Closing connection")
		client.connection.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			zap.S().Debug("Setting connection write deadline")
			err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err != nil {
				zap.L().Error("Error setting write deadline", zap.Error(err))
				return
			}

			if !ok {
				zap.S().Info("Sending message was not OK, closing connection..")
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			zap.S().Debug("Defining next writer")

			writer, err := client.connection.NextWriter(websocket.TextMessage)

			if err != nil {
				zap.L().Error("Error defining websocket writer", zap.Error(err))
				return
			}

			zap.S().Debugw("Writing message", "message", string(message))

			_, err = writer.Write(message)

			if err != nil {
				zap.L().Error("Error writing the actual message", zap.Error(err))
				return
			}

			zap.S().Info("Sending queued messages")

			// Handle queued messages
			for i := 0; i < len(client.send); i++ {
				_, err = writer.Write(newLine)

				if err != nil {
					zap.L().Error("Error writing newline", zap.Error(err))
					return
				}

				message, ok := <-client.send

				if !ok {
					zap.S().Info("Sending queued message was not OK, closing connection..")
					client.connection.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				zap.S().Debugw("Writing message", "message", string(message))
				_, err = writer.Write(message)

				if err != nil {
					zap.L().Error("Error writing message", zap.Error(err))
					return
				}
			}

			if err := writer.Close(); err != nil {
				zap.L().Error("Error closing writer", zap.Error(err))
				return
			}
		case <-ticker.C:
			zap.S().Info("Sending Ping Message")

			err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if err != nil {
				zap.L().Error("Error setting write deadline", zap.Error(err))
				return
			}

			if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				zap.L().Error("Error sending ping message", zap.Error(err))
				return
			}
		}
	}
}

func HandleNewLogLine(client *Client, message LogMessage) {
	zap.S().Debugw(
		"Sending new log line message",
		"message", string(message.Line),
		"clientId", client.id,
	)

	client.hub.broadcast <- struct {
		message  []byte
		clientId string
	}{
		message:  []byte(message.Line),
		clientId: client.id,
	}
}

func HandleIdentityMessage(client *Client, message Message, payload IdentityMessage) {
	zap.S().Debugw(
		"Handling identity message",
		"message", string(message.Event),
		"clientId", client.id,
	)

	var updateId string

	// In case this is a broadcaster peer
	if payload.Broadcaster {
		zap.S().Debugw(
			"Preparing broadcaster client",
			"updateId", client.id,
			"broadcaster", payload.Broadcaster,
		)

		updateId = client.id
		client.id = message.Id
		client.broadcaster = payload.Broadcaster
		client.peerId = ""
	} else {
		zap.S().Debugw(
			"Preparing remote client",
			"updateId", client.id,
			"broadcaster", payload.Broadcaster,
		)

		if payload.PeerId == "" {
			zap.S().Warn("Remote client identity was sent with empty peerId, discarding...")
			return
		}

		updateId = client.id
		client.peerId = payload.PeerId
	}

	zap.S().Debug("Setting client as active")

	client.active = true

	client.hub.update <- struct {
		id     string
		client *Client
	}{updateId, client}

	zap.S().Debugw(
		"Update request was sent",
		"id", client.id,
		"peerId", client.peerId,
		"active", client.active,
		"broadcaster", client.broadcaster,
	)
}

func HandleMessage(client *Client) (Message, error) {
	zap.S().Debugw(
		"Handling client incoming messages",
		"id", client.id,
	)

	var message Message
	err := client.connection.ReadJSON(&message)

	if err != nil {
		zap.L().Error("Error occurred while reading incoming JSON message", zap.Error(err))

		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			zap.L().Warn("Unexpected websocket close, peer is disconnected, ignoring message...")
		}

		return Message{}, err
	}

	switch message.Event {
	case EVENT_IDENTITY:
		zap.S().Debug("Incoming identity event")
		identityMessage := IdentityMessage{}
		data, err := json.Marshal(message.Payload)

		if err != nil {
			zap.L().Error("Unexpected error while marshaling payload", zap.Error(err))
			return Message{}, err
		}

		zap.S().Debugw(
			"Payload was marshaled",
			"payload", string(data),
		)

		err = json.Unmarshal([]byte(data), &identityMessage)

		if err != nil {
			zap.L().Error("Unexpected error while unmarshaling payload", zap.Error(err))
			return Message{}, err
		}

		HandleIdentityMessage(client, message, identityMessage)

	case EVENT_LOG_LINE:
		if !client.active {
			zap.L().Warn("Client is not active yet, ignoring message")
			return Message{}, errors.New("Client is not active yet, ignoring messages")
		}

		zap.S().Debug("Incoming log_line event, preparing log message")

		logMessage := LogMessage{}
		data, err := json.Marshal(message.Payload)

		if err != nil {
			zap.L().Error("Unexpected error while marshaling payload", zap.Error(err))
			return Message{}, err
		}

		zap.S().Debugw(
			"Payload was marshaled",
			"payload", string(data),
		)

		err = json.Unmarshal([]byte(data), &logMessage)

		if err != nil {
			zap.L().Error("Unexpected error while unmarshaling payload", zap.Error(err))
			return Message{}, err
		}

		HandleNewLogLine(client, logMessage)
	}

	return message, nil
}
