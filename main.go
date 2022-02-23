package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"

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

	go ScanFile(input)
	go HandleIncomingMessages(connection)

	// Main loop of the client
	// Here we send & receive packets
	for {
		select {
		case line := <-input:
			// log.Printf("[%s] %s\n", clientId, line)
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
