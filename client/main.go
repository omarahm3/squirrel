package client

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/utils"
	"go.uber.org/zap"
)

var interrupt chan os.Signal
var options *ClientOptions

func Main() {
	options = InitOptions()
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate
	input := make(chan string)

	utils.InitLogging(utils.LoggerOptions{
		Env:         options.Env,
		LogLevel:    options.LogLevel,
		LogFileName: ".squirrel.log",
	})

	defer func() {
		_ = zap.L().Sync()
		_ = zap.S().Sync()
	}()

	clientId := utils.GenerateUUID()

	zap.S().Debug("Client ID was generated: ", clientId)

	fmt.Printf("Link: [ %s/client/%s ]\n", options.Domain.Public, clientId)

	signal.Notify(interrupt, os.Interrupt)

	connection := InitClient()

	defer connection.Close()

	SendIdentity(connection, clientId)

	go ScanFile(input)
	go HandleIncomingMessages(connection)

	// Main loop of the client
	// Here we send & receive packets
	for {
		select {
		case line := <-input:
			// log.Printf("[%s] %s\n", clientId, line)
			err := connection.WriteJSON(Message{
				Id:    clientId,
				Event: "log_line",
				Payload: LogMessage{
					Line: line,
				},
			})

			if err != nil {
				zap.S().Error("Error during sending message to websocket:", zap.Error(err))
				return
			}
		case <-interrupt:
			zap.S().Info("Received SIGINT interrupt signal. Closing all pending connections")
			HandleWebsocketClose(connection)
			return
		}
	}
}

func SendIdentity(connection *websocket.Conn, clientId string) {
	message := Message{
		Id:    clientId,
		Event: "identity",
		Payload: IdentityMessage{
			PeerId:      "",
			Broadcaster: true,
			Subscriber:  false,
		},
	}

	zap.L().Info("Sending client identity: ", zap.Object("message", message))

	err := connection.WriteJSON(message)

	zap.S().Info("Identity message sent")

	if err != nil {
		zap.S().Error("Errer sending identity message", zap.Error(err))
		return
	}
}

func ScanFile(input chan string) {
	zap.S().Debug("Scanning log file")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		input <- text
	}

	if err := scanner.Err(); err != nil {
		zap.S().Error("Error scanning file", zap.Error(err))
		return
	}
}
