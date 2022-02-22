package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
)

var interrupt chan os.Signal

func main() {
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate
	input := make(chan string)

	signal.Notify(interrupt, os.Interrupt)

	connection := InitClient(input)
	defer connection.Close()

  go ScanFile(input)

	// Main loop of the client
	// Here we send & receive packets
	for {
		select {
		case line := <-input:
			log.Println(line)
			SendMessage(connection, line)

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
		log.Println(err)
	}
}
