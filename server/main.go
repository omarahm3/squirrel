package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/omarahm3/squirrel/utils"
	"go.uber.org/zap"
)

var (
	options *ServerOptions
	hub     *Hub
	server  *gin.Engine
)

func printOptions() {
	zap.S().Warnw(
		"Server started with these options",
		"Env", options.Env,
		"Public Domain", options.Domain.Public,
		"Websocket Domain", options.Domain.Websocket,
		"Port", options.Port,
		"Log Level", options.LogLevel.String(),
		"Read Buffer Size", options.ReadBufferSize,
		"Write Buffer Size", options.WriteBufferSize,
	)
}

func Main() {
	options = InitOptions()

	if options.Env != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	utils.InitLogging(utils.LoggerOptions{
		Env:         options.Env,
		LogLevel:    options.LogLevel,
		LogFileName: ".server.squirrel.log",
	})

	printOptions()

	// Sync both loggers since they're all used
	defer func() {
		_ = zap.L().Sync()
		_ = zap.S().Sync()
	}()

	server = gin.Default()

	zap.S().Debug("Prepared server default")

	hub = NewHub()

	zap.S().Debug("Created clients hub")

	go hub.Run()

	zap.S().Debug("Loading server HTML files")

	server.LoadHTMLFiles("./server/view/index.html")

	InitHttpServer()

	zap.S().Debugf("Running server on port [%d]\n", options.Port)

	err := server.Run(fmt.Sprintf(":%d", options.Port))

	if err != nil {
		utils.FatalError("Error while running server", err)
	}

	fmt.Printf("Server is running on http://localhost:%d", options.Port)
}
