package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap"
)

func InitHttpServer() {
	zap.S().Debug("Initializing server routes")

	server.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})

	server.GET("/ws", func(context *gin.Context) {
		WebsocketHandler(context.Request, context.Writer)
	})

	server.GET("/client/:clientId", SubscriberView)
}

func WebsocketHandler(r *http.Request, w http.ResponseWriter) {
	zap.S().Info("Handling websocket upgrade request")

	var wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  options.ReadBufferSize,
		WriteBufferSize: options.WriteBufferSize,
	}

	connection, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		zap.L().Error("Error upgrading websocket request, ignoring", zap.Error(err))
		return
	}

	zap.S().Info("Websocket connection was successful")

	client := &Client{
		id:          common.GenerateUUID(),
		connection:  connection,
		hub:         hub,
		broadcaster: false,
		send:        make(chan []byte, 256),
	}

	zap.S().Infow("Initialized new client", "clientId", client.id)

	client.hub.register <- client

	go client.ReadPump()
	go client.WritePump()
}

func SubscriberView(context *gin.Context) {
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
		"domain":   options.Domain.Websocket,
	})
}
