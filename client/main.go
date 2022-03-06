package client

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/atotto/clipboard"
	"github.com/gorilla/websocket"
	"github.com/inancgumus/screen"
	"github.com/omarahm3/squirrel/common"
	"go.uber.org/zap"
)

type ControllerMessage struct {
	Message    string
	Error      error
	Connection *websocket.Conn
}

var (
	interrupt  chan os.Signal
	options    *ClientOptions
	clientId   string
	controller = make(chan int)
	events     = make(chan string)
	input      = make(chan string)
)

const (
	EVENT_IDENTITY       = "identity"
	EVENT_SUBSCRIBER_ACK = "subscriber_ack"
)

func isStdin() bool {
	stat, err := os.Stdin.Stat()

	if err != nil {
		fmt.Println("Couldn't check STDIN: ", err)
		os.Exit(1)
	}

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func Main() {
	options = InitOptions()

	if !options.Listen && !isStdin() {
		fmt.Println("Nothing is being read, you should pipe something to stdin of this command")
		return
	}

	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to gracefully terminate

	common.InitLogging(common.LoggerOptions{
		Env:         options.Env,
		LogLevel:    options.LogLevel,
		LogFileName: ".squirrel.log",
	})

	defer func() {
		_ = zap.L().Sync()
		_ = zap.S().Sync()
	}()

	clientId = common.GenerateUUID()

	zap.S().Debug("Client ID was generated: ", clientId)

	signal.Notify(interrupt, os.Interrupt)

	connection := InitClient()

	defer connection.Close()

	SendIdentity(connection, clientId)

	if !options.Listen {
		link := fmt.Sprintf("%s/client/%s", options.Domain.Public, clientId)

		fmt.Printf("➜ ID: [ %s ]\n", clientId)
		fmt.Printf("➜ Link: [ %s ]\n", link)

		if options.UrlClipboard {
			err := clipboard.WriteAll(link)

			if err != nil {
				zap.S().Warnw("Error occurred while writing link to clipboard", "error", zap.Error(err))
			} else {
				fmt.Println("➜ Url is copied to your clipboard")
				fmt.Println("📢 Squirrel is waiting for listeners to begin piping stdout...")
			}
		}
	}

	go HandleEvents()
	go HandleSendEvents(connection)
	go HandleIncomingMessages(connection)

	// Main CLI loop
	for {
		select {
		case <-interrupt:
			zap.S().Info("Received SIGINT interrupt signal. Closing all pending connections")
			return
		case <-controller:
			return
		}
	}
}

func HandleSendEvents(connection *websocket.Conn) {
	defer func() {
		connection.Close()
		zap.S().Info("Client connection closed")
	}()

	// Here we receive packets
	for {
		line := <-input
		err := connection.WriteJSON(common.Message{
			Id:    clientId,
			Event: "log_line",
			Payload: common.LogMessage{
				Line: line,
			},
		})

		if err != nil {
			zap.S().Error("Error during sending message to websocket:", zap.Error(err))
			return
		}
	}
}

func SendIdentity(connection *websocket.Conn, clientId string) {
	var peerId string
	var subscriber bool
	broadcaster := true

	if options.PeerId != "" && options.Listen {
		peerId = options.PeerId
		subscriber = true
		broadcaster = false
	}

	message := common.Message{
		Id:    clientId,
		Event: EVENT_IDENTITY,
		Payload: common.IdentityMessage{
			PeerId:      peerId,
			Broadcaster: broadcaster,
			Subscriber:  subscriber,
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

func HandleEvents() {
	for {
		event := <-events

		switch event {
		case EVENT_SUBSCRIBER_ACK:
			screen.Clear()
			screen.MoveTopLeft()
			go ScanFile()
		}
	}
}

func ScanFile() {
	zap.S().Debug("Scanning log file")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		if options.Output {
			fmt.Println(text)
		}

		input <- text
	}

	if err := scanner.Err(); err != nil {
		zap.S().Error("Error scanning file", zap.Error(err))
		return
	}
}
