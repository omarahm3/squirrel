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
	log.Printf("Link: [ http://localhost:3000/client/%s ]\n", clientId)

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
				log.Println("Error during sending message to websocket:", err)
				return
			}
		case <-interrupt:
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")
			HandleWebsocketClose(connection)
			return
		}
	}
}

func SendIdentity(connection *websocket.Conn, clientId string) {
	err := connection.WriteJSON(Message{
		Id:    clientId,
		Event: "identity",
		Payload: IdentityMessage{
      Local: true,
		},
	})

	if err != nil {
		log.Fatal("Error sending identity message", err)
		return
	}
}

func ScanFile(input chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		input <- text
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error scanning file", err)
		return
	}
}
