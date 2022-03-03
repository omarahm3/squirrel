package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type Message struct {
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
}

func (m Message) MarshalPayload() ([]byte, error) {
	data, err := json.Marshal(m.Payload)

	if err != nil {
		zap.L().Error("Unexpected error while marshaling payload", zap.Error(err))
		return []byte{}, err
	}

	zap.S().Debugw(
		"Payload was marshaled",
		"payload", string(data),
	)

	return data, nil
}

func (m Message) ToLogMessage() (LogMessage, error) {
	data, err := m.MarshalPayload()

	if err != nil {
		return LogMessage{}, err
	}

	logMessage := LogMessage{}

	err = json.Unmarshal([]byte(data), &logMessage)

	if err != nil {
		zap.L().Error("Unexpected error while unmarshaling payload", zap.Error(err))
		return LogMessage{}, err
	}

	return logMessage, nil
}

func (m Message) ToIdentityMessage() (IdentityMessage, error) {
	data, err := m.MarshalPayload()

	if err != nil {
		zap.L().Error("Unexpected error while marshaling payload", zap.Error(err))
		return IdentityMessage{}, err
	}

	zap.S().Debugw(
		"Payload was marshaled",
		"payload", string(data),
	)

	identityMessage := IdentityMessage{}
	err = json.Unmarshal([]byte(data), &identityMessage)

	if err != nil {
		zap.L().Error("Unexpected error while unmarshaling payload", zap.Error(err))
		return IdentityMessage{}, err
	}

	return identityMessage, nil
}

type LogMessage struct {
	Line string `json:"line"`
}

func (message LogMessage) Handle(client *Client) {
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

type IdentityMessage struct {
	PeerId      string `json:"peerId"`
	Broadcaster bool   `json:"broadcaster"`
	Subscriber  bool   `json:"subscriber"`
}

func (payload IdentityMessage) Handle(client *Client, message Message) error {
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

func HandleMessage(client *Client, message Message) (Message, error) {
	switch message.Event {
	case EVENT_IDENTITY:
		identityMessage, err := message.ToIdentityMessage()

		if err != nil {
			return Message{}, err
		}

		err = identityMessage.Handle(client, message)

		if err != nil {
			return Message{}, err
		}

	case EVENT_LOG_LINE:
		if !client.active {
			zap.L().Warn("Client is not active yet, ignoring message")
			return Message{}, errors.New("Client is not active yet, ignoring messages")
		}

		zap.S().Debug("Incoming log_line event, preparing log message")

		logMessage, err := message.ToLogMessage()

		if err != nil {
			return Message{}, err
		}

		logMessage.Handle(client)
	}

	return message, nil
}
