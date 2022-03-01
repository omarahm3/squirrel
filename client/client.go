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

// Needed to receive server events
// Right now we do nothing, but its here to avoid errors on the protocol
func HandleIncomingMessages(connection *websocket.Conn) {
	defer func() {
		connection.Close()
		zap.S().Info("Client connection closed")
	}()

	for {
		_, message, err := connection.ReadMessage()

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
