package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const WEBSOCKET_URL string = "ws://localhost:3000/ws"

var done chan interface{}
var interrupt chan os.Signal

func InitClient(input chan string) {
	done = make(chan interface{})    // Channel to indicate receiveHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate

	signal.Notify(interrupt, os.Interrupt)

	connection, _, err := websocket.DefaultDialer.Dial(WEBSOCKET_URL, nil)

	if err != nil {
		log.Fatal("Error connecting to websocket server:", err)
	}

	defer connection.Close()

	go ReceiveHandler(connection)
  go ScanFile(input)

	// Main loop of the client
	// Here we send & receive packets
	for {
		select {
    case line := <- input:
      log.Println(line)
      err := connection.WriteMessage(websocket.TextMessage, []byte(line))

      if err != nil {
        log.Println("Error during sending message to websocket:", err)
        return
      }

		case <-interrupt:
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

			if err != nil {
				log.Fatal("Error during closing websocket:", err)
				return
			}

			select {
			// case <-done:
			// 	log.Println("Receiver channel closed! Exiting...")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel, Exiting...")
			}

			return
		}
	}
}

func ReceiveHandler(connection *websocket.Conn) {
	defer close(done)

	for {
		_, message, err := connection.ReadMessage()

		if err != nil {
      if !websocket.IsCloseError(err, 1000) {
        log.Println("Error receiving message:", err)
      }

			return
		}

		log.Printf("Received: %s\n", message)
	}
}
