package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

func (m Message) ToSubscriberConnectedMessage() (SubscriberConnectedMessage, error) {
	data, err := m.MarshalPayload()

	if err != nil {
		zap.L().Error("Unexpected error while marshaling payload", zap.Error(err))
		return SubscriberConnectedMessage{}, err
	}

	zap.S().Debugw(
		"Payload was marshaled",
		"payload", string(data),
	)

	message := SubscriberConnectedMessage{}
	err = json.Unmarshal([]byte(data), &message)

	if err != nil {
		zap.L().Error("Unexpected error while unmarshaling payload", zap.Error(err))
		return SubscriberConnectedMessage{}, err
	}

	return message, nil
}

func NewMessage(message []byte) (Message, error) {
	var m Message

	err := json.Unmarshal([]byte(message), &m)

	if err != nil {
		return Message{}, err
	}

	return m, nil
}

func (message Message) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", message.Id)
	encoder.AddString("Event", message.Event)

	data, err := json.Marshal(message.Payload)

	if err != nil {
		zap.S().Error("Unexpected error while marshaling payload: ", err, message.Payload)
		return err
	}

	encoder.AddString("payload", string(data))
	return nil
}

func InitClient() *websocket.Conn {
	zap.S().Debug("Initiating websocket client")

	connection, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", options.Domain.Websocket), nil)

	if err != nil {
		zap.S().Error("Error connecting to websocket server: ", err)
		os.Exit(1)
	}

	zap.S().Debug("Websocket connection was successful")

	return connection
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// Needed to receive server events
// Right now we do nothing, but its here to avoid errors on the protocol
func HandleIncomingMessages(connection *websocket.Conn) {
	defer func() {
		connection.Close()
		zap.S().Info("Client connection closed")
	}()

	for {
		_, message, err := connection.ReadMessage()

		if isJSON(string(message)) {
			jsonMessage, err := NewMessage(message)

			if err != nil {
				break
			}

			if jsonMessage.Event == "subscriber_ack" {
				m, err := jsonMessage.ToSubscriberConnectedMessage()

				if err != nil {
					break
				}

				if m.Connected {
					fmt.Println("Subscriber is connected")
				}
			}
		}

		if options.Listen && options.PeerId != "" {
			fmt.Println(string(message))
		}

		if err != nil {
			HandleWebsocketClose(ControllerMessage{
				Error:      err,
				Connection: connection,
				Message:    "Error while reading incoming message",
			})
			break
		}
	}
}

func HandleWebsocketClose(message ControllerMessage) {
	if websocket.IsUnexpectedCloseError(message.Error, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseMessage) {
		controller <- 1
	} else {
		controller <- 0
		zap.L().Error("Error occurred, closing websocket connection", zap.Error(message.Error))
	}
}
