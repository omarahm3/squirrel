package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omarahm3/live-logs/utils"
	"go.uber.org/zap"
)

const SERVER_PORT = 3000

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
		local:      false,
		send:       make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.ReadPump()
  go client.WritePump()
}

func main() {
  utils.InitLogging()

	server := gin.Default()

  zap.S().Debug("Prepared server default")

	hub := NewHub()

  zap.S().Debug("Created clients hub")

	go hub.Run()

  zap.S().Debug("Loading server HTML files")

	server.LoadHTMLFiles("./view")

  initRoutes(server, hub)

  zap.S().Debugf("Running server on port [%d]\n", SERVER_PORT)

  err := server.Run(fmt.Sprintf(":%d", SERVER_PORT))

	if err != nil {
    utils.FatalError("Error while running server", err)
	}

	fmt.Printf("Server is running on http://localhost:%d", SERVER_PORT)
}

func initRoutes(server *gin.Engine, hub *Hub) {
  zap.S().Debug("Initializing server routes")

	server.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})

	server.GET("/ws", func(context *gin.Context) {
		wsHandler(hub, context.Request, context.Writer)
	})

	server.GET("/client/:clientId", func(context *gin.Context) {
		clientId := context.Param("clientId")

    zap.S().Debugf("Incoming request to subscribe to client ID: [%s]\n", clientId)

		if clientId == "" {
      zap.S().Debug("Client ID is empty ignoring")
			context.String(400, "Cannot be empty")
			return
		}

		if _, ok := hub.clients[clientId]; !ok {
      zap.S().Debugf("Client ID: [%s] doesn't exist on the hub\n", clientId)
			context.String(404, "Client not found")
			return
		}

    zap.S().Debugf("Client ID: [%s] was found on hub\n", clientId)

		context.HTML(200, "index.html", gin.H{
			"clientId": clientId,
		})
	})
}
