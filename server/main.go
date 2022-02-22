package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omarahm3/live-logs/utils"
)

type LogMessage struct {
	Id   string `json:"id"`
	Line string `json:"line"`
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(hub *Hub, r *http.Request, w http.ResponseWriter) {
	connection, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	client := &Client{
		id:         utils.GenerateUUID(),
		connection: connection,
		hub:        hub,
		send:       make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.ReadPump()
}

func main() {
	server := gin.Default()
	hub := NewHub()

	go hub.Run()

	server.LoadHTMLFiles("./view/index.html")

	server.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})

	server.GET("/ws", func(c *gin.Context) {
		wsHandler(hub, c.Request, c.Writer)
	})

	err := server.Run(":3000")

	if err != nil {
		panic(fmt.Sprintf("Failed to start server - Error %v", err))
	}

	fmt.Println("Server is running on http://localhost:3000")
}
