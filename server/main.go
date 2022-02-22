package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type LogMessage struct {
	Id   string `json:"id"`
	Line string `json:"line"`
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(r *http.Request, w http.ResponseWriter) {
	connection, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	for {
    var message LogMessage
		err := connection.ReadJSON(&message)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Unexpected server close: ", err)
			}

			break
		}

		fmt.Println(message)

		err = connection.WriteJSON(message)

		if err != nil {
			break
		}
	}
}

func main() {
	server := gin.Default()

	server.LoadHTMLFiles("./view/index.html")

	server.GET("/", func(context *gin.Context) {
		context.HTML(200, "index.html", nil)
	})

	server.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Request, c.Writer)
	})

	err := server.Run(":3000")

	if err != nil {
		panic(fmt.Sprintf("Failed to start server - Error %v", err))
	}

	fmt.Println("Server is running on http://localhost:3000")
}
