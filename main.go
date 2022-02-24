package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/live-logs/utils"
	"go.uber.org/zap"
)

var interrupt chan os.Signal

func main() {
  interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate
	input := make(chan string)

	utils.HandleLogLevel()

  defer func() {
    if err := zap.S().Sync(); err != nil {
      fmt.Println("Error syncing zap:", err)
      os.Exit(1)
    }
  }()

	clientId := utils.GenerateUUID()

	zap.S().Debug("Client ID was generated: ", clientId)

	fmt.Printf("Link: [ http://localhost:3000/client/%s ]\n", clientId)

	signal.Notify(interrupt, os.Interrupt)

	connection := InitClient(input)

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
			Local: true,
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
