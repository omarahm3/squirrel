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
	PeerId string `json:"peerId"`
	Local  bool   `json:"local"`
}

type Message struct {
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
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

func InitClient(input chan string) *websocket.Conn {
	zap.S().Debug("Initiating websocket client")

	connection, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://%s/ws", options.Domain), nil)

	if err != nil {
		zap.S().Error("Error connecting to websocket server: ", err)
		os.Exit(1)
	}

	zap.S().Debug("Websocket connection was successful")

	return connection
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

		zap.S().Debug("Incoming message: ", string(message))

		if err != nil {
			zap.L().Error("Error while reading incoming message", zap.Error(err))
			break
		}
	}
}

func HandleWebsocketClose(connection *websocket.Conn) {
	zap.L().Info("Closing websocket connection")

	err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	if err != nil {
		zap.L().Error("Error during closing websocket:", zap.Error(err))
		return
	}
}
