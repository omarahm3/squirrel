package server

import (
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

func (client *Client) IsActiveBroadcaster() bool {
	return client.broadcaster && client.active
}

func (client *Client) IsActiveSubscriber() bool {
	return client.subscriber && client.active && client.peerId != ""
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
		zap.S().Debug("Received Pong message")
		client.connection.SetReadDeadline(time.Now().Add(PONG_WAIT))
		return nil
	})

	for {
		message, err := ReadIncomingMessage(client)

		if err != nil {
			zap.L().Error("Error handling message, disconnecting peer", zap.Error(err), zap.String("peerId", client.id))
			return
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
				_, err = writer.Write([]byte{'\n'})

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

func ReadIncomingMessage(client *Client) (Message, error) {
	zap.S().Debugw(
		"Handling client incoming messages",
		"id", client.id,
	)

	var message Message

	zap.S().Debugw(
		"Reading message of client",
		"client", client.id,
		"broadcaster", client.broadcaster,
		"subscriber", client.subscriber,
	)

	err := client.connection.ReadJSON(&message)

	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
			zap.L().Warn("Unexpected websocket close, peer is disconnected, ignoring message...")
		}

		return Message{}, err
	}

	return HandleMessage(client, message)
}
