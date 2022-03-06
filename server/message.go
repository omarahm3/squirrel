package server

import (
	"errors"
	"fmt"

	"github.com/omarahm3/squirrel/common"
	"go.uber.org/zap"
)

func HandleLogMessage(message common.LogMessage, client *Client) {
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

func HandleIdentityMessage(payload common.IdentityMessage, client *Client, message common.Message) error {
	zap.S().Debugw(
		"Handling identity message",
		"event", string(message.Event),
		"clientId", client.id,
		"broadcaster", payload.Broadcaster,
		"subscriber", payload.Subscriber,
	)

	var updateId string

	// In case this is a broadcaster peer
	if payload.Broadcaster {
		zap.S().Debugw(
			"Preparing broadcaster client",
			"updateId", client.id,
			"broadcaster", payload.Broadcaster,
			"subscriber", payload.Subscriber,
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
			"subscriber", payload.Subscriber,
		)

		if payload.PeerId == "" {
			zap.S().Warn("Remote client identity was sent with empty peerId, discarding...")
			return nil
		}

		updateId = client.id
		client.peerId = payload.PeerId
		client.subscriber = payload.Subscriber
	}

	zap.S().Debug("Setting client as active")

	client.active = true

	client.hub.update <- struct {
		id     string
		client *Client
	}{updateId, client}

	if client.IsActiveSubscriber() {
		if _, ok := client.hub.clients[client.peerId]; !ok {
			return fmt.Errorf("Client ID: [%s] doesn't exist on the hub", client.peerId)
		}
	}

	zap.S().Debugw(
		"Update request was sent",
		"id", client.id,
		"peerId", client.peerId,
		"active", client.active,
		"broadcaster", client.broadcaster,
		"subscriber", client.subscriber,
	)

	return nil
}

func HandleMessage(client *Client, message common.Message) (common.Message, error) {
	switch message.Event {
	case EVENT_IDENTITY:
		identityMessage, err := message.ToIdentityMessage()

		if err != nil {
			return common.Message{}, err
		}

		err = HandleIdentityMessage(identityMessage, client, message)

		if err != nil {
			return common.Message{}, err
		}

	case EVENT_LOG_LINE:
		if !client.active {
			zap.L().Warn("Client is not active yet, ignoring message")
			return common.Message{}, errors.New("Client is not active yet, ignoring messages")
		}

		zap.S().Debug("Incoming log_line event, preparing log message")

		logMessage, err := message.ToLogMessage()

		if err != nil {
			return common.Message{}, err
		}

		HandleLogMessage(logMessage, client)
	}

	return message, nil
}
