package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/live-logs/utils"
)

var interrupt chan os.Signal

func main() {
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate
	input := make(chan string)
	clientId := utils.GenerateUUID()

	signal.Notify(interrupt, os.Interrupt)

	connection := InitClient(input)
	defer connection.Close()

	go ScanFile(input)
  go incomingMessages(connection)

	// Main loop of the client
	// Here we send & receive packets
	for {
		select {
		case line := <-input:
			log.Printf("[%s] %s\n", clientId, line)
			SendMessage(connection, Message{
				Id:    clientId,
				Local: true,
				Event: "log_line",
				Payload: LogMessage{
					Line: line,
				},
			})
		case <-interrupt:
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")
			HandleWebsocketClose(connection)
			return
		}
	}
}

// Needed to receive server events
// Right now we do nothing, but its here to avoid errors on the protocol
func incomingMessages(connection *websocket.Conn) {
	defer func() {
    connection.Close()
	}()

	for {
		_, _, err := connection.ReadMessage()

		if err != nil {
			break
		}
	}
}

func ScanFile(input chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		input <- text
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
