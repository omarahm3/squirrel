package client

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap"
)

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

func handleIncomingJSONMessages(message []byte) error {
	jsonMessage, err := common.NewMessageFromString(message)

	if err != nil {
		return err
	}

	if jsonMessage.Event == EVENT_SUBSCRIBER_ACK {
		m, err := jsonMessage.ToSubscriberConnectedMessage()

		if err != nil {
			return err
		}

		if m.Connected {
			select {
			case events <- EVENT_SUBSCRIBER_ACK:
				zap.S().Debug("sent subscriber_ack event")
			default:
				zap.S().Debug("no events were sent")
			}
		}
	}

	return nil
}

func readIncomingMessages(connection *websocket.Conn) bool {
	_, message, err := connection.ReadMessage()

	if common.IsJSON(string(message)) {
		err := handleIncomingJSONMessages(message)

		if err != nil {
			return false
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
		return false
	}

	return true
}

func HandleIncomingMessages(connection *websocket.Conn) {
	defer func() {
		connection.Close()
		zap.S().Info("Client connection closed")
	}()

	for {
		result := readIncomingMessages(connection)

		if !result {
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
