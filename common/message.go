package common

import (
	"encoding/json"

	"go.uber.org/zap"
)

type Message struct {
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
}

type LogMessage struct {
	Line string `json:"line"`
}

type IdentityMessage struct {
	PeerId      string `json:"peerId"`
	Broadcaster bool   `json:"broadcaster"`
	Subscriber  bool   `json:"subscriber"`
}

type SubscriberConnectedMessage struct {
	Connected bool `json:"connected"`
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

func (m Message) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)

	if err != nil {
		zap.L().Error("Unexpected error while marshaling message", zap.Error(err))
		return []byte{}, err
	}

	zap.S().Debugw(
		"Message was marshaled",
		"message", string(data),
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

func (m SubscriberConnectedMessage) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)

	if err != nil {
		zap.L().Error("Unexpected error while marshaling SubscriberConnectedMessage", zap.Error(err))
		return []byte{}, err
	}

	return data, nil
}
