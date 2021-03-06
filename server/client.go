package server

import (
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap"
)

const (
	WRITE_WAIT           = 10 * time.Second
	PONG_WAIT            = 60 * time.Second
	PING_PERIOD          = (PONG_WAIT * 9) / 10
	EVENT_IDENTITY       = "identity"
	EVENT_LOG_LINE       = "log_line"
	EVENT_SUBSCRIBER_ACK = "subscriber_ack"
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

func (client *Client) ReadIncomingMessage() (common.Message, error) {
	zap.S().Debugw(
		"Handling client incoming messages",
		"id", client.id,
	)

	var message common.Message

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

		return common.Message{}, err
	}

	return HandleMessage(client, message)
}

func (client *Client) writeMessage(message []byte) (io.WriteCloser, error) {
	zap.S().Debug("Setting connection write deadline")
	err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

	if err != nil {
		zap.L().Error("Error setting write deadline", zap.Error(err))
		return nil, err
	}

	writer, err := client.connection.NextWriter(websocket.TextMessage)

	if err != nil {
		zap.L().Error("Error defining websocket writer", zap.Error(err))
		return nil, err
	}

	zap.S().Debugw("Writing message", "message", string(message))

	_, err = writer.Write(message)

	if err != nil {
		zap.L().Error("Error writing the actual message", zap.Error(err))
		return nil, err
	}

	return writer, nil
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
		_, err := client.ReadIncomingMessage()

		if err != nil {
			zap.L().Error("Error handling message, disconnecting peer", zap.Error(err), zap.String("peerId", client.id))
			return
		}

		if client.IsActiveSubscriber() {
			zap.S().Debugw(
				"Client is active subscriber",
				"peerId", client.peerId,
				"id", client.id,
			)

			ackPayload := &common.SubscriberConnectedMessage{
				Connected: true,
			}

			ackMessage := common.Message{
				Id:      "",
				Payload: ackPayload,
				Event:   EVENT_SUBSCRIBER_ACK,
			}

			message, err := ackMessage.Marshal()

			if err != nil {
				return
			}

			client.hub.send <- struct {
				message  []byte
				clientId string
			}{
				message:  message,
				clientId: client.peerId,
			}
		}
	}
}

func writeQueuedMessages(client *Client, writer io.WriteCloser) error {
	zap.S().Info("Sending queued messages")

	// Handle queued messages
	for i := 0; i < len(client.send); i++ {
		_, err := writer.Write([]byte{'\n'})

		if err != nil {
			zap.L().Error("Error writing newline", zap.Error(err))
			return err
		}

		message, ok := <-client.send

		if !ok {
			zap.S().Info("Sending queued message was not OK, closing connection..")
			client.connection.WriteMessage(websocket.CloseMessage, []byte{})
			return err
		}

		zap.S().Debugw("Writing message", "message", string(message))
		_, err = writer.Write(message)

		if err != nil {
			zap.L().Error("Error writing message", zap.Error(err))
			return err
		}
	}

	return nil
}

func sendPingMessage(client *Client) error {
	zap.S().Info("Sending Ping Message")

	err := client.connection.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

	if err != nil {
		zap.L().Error("Error setting write deadline", zap.Error(err))
		return err
	}

	if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
		zap.L().Error("Error sending ping message", zap.Error(err))
		return err
	}

	return nil
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
			if !ok {
				zap.S().Info("Sending message was not OK, closing connection..")
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := client.writeMessage(message)

			if err != nil {
				return
			}

			err = writeQueuedMessages(client, writer)

			if err != nil {
				return
			}

			if err := writer.Close(); err != nil {
				zap.L().Error("Error closing writer", zap.Error(err))
				return
			}
		case <-ticker.C:
			err := sendPingMessage(client)

			if err != nil {
				return
			}
		}
	}
}
