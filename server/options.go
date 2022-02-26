package server

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/omarahm3/squirrel/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ServerOptions struct {
	Env             string
	Domain          string
	Port            int
	LogLevel        zapcore.Level
	ReadBufferSize  int
	WriteBufferSize int
}

const (
	DEFAULT_ENVIRONMENT       = "prod"
	DEFAULT_DOMAIN            = "localhost:3000"
	DEFAULT_PORT              = "3000"
	DEFAULT_LOG_LEVEL         = "error"
	DEFAULT_READ_BUFFER_SIZE  = "0"
	DEFAULT_WRITE_BUFFER_SIZE = "0"
)

func fprintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func winningDefault(envVariable string, defaultVariable string) string {
	if envVariable == "" {
		return defaultVariable
	}
	return envVariable
}

func toInt(value string) int {
	intVal, err := strconv.Atoi(value)

	if err != nil {
		fmt.Println("Error converting value to int", err)
		os.Exit(1)
	}

	return intVal
}

func handleLogging(loglevel string) zapcore.Level {
	level, err := zapcore.ParseLevel(loglevel)

	if err != nil {
		fmt.Println("Invalid log level returned, setting default log level, Error: ", err)
		level = zap.ErrorLevel
	}

	return level
}

func InitOptions() *ServerOptions {
	flag.Usage = func() {
		fprintf("Usage of %s:\n", os.Args[0])
		fprintf(" %s [options]\n", os.Args[0])
		fprintf("Options:\n")
		flag.PrintDefaults()
		fprintf("%s configures server run.\n", os.Args[0])
	}

	var (
		env             string = utils.GetEnvVariable("APP_ENV")
		domain          string = utils.GetEnvVariable("DOMAIN")
		port            string = utils.GetEnvVariable("PORT")
		loglevel        string = utils.GetEnvVariable("LOG_LEVEL")
		readBufferSize  string = utils.GetEnvVariable("READ_BUFFER_SIZE")
		writeBufferSize string = utils.GetEnvVariable("WRITE_BUFFER_SIZE")
	)

	flag.StringVar(&env, "env", winningDefault(utils.GetEnvVariable("APP_ENV"), DEFAULT_ENVIRONMENT), "Server environment (prod|dev)")
	flag.StringVar(&domain, "domain", winningDefault(utils.GetEnvVariable("DOMAIN"), DEFAULT_DOMAIN), "Server domain")
	flag.StringVar(&loglevel, "log", winningDefault(utils.GetEnvVariable("LOG_LEVEL"), DEFAULT_LOG_LEVEL), "Log level")
	flag.StringVar(&port, "port", winningDefault(utils.GetEnvVariable("PORT"), DEFAULT_PORT), "Server port")
	flag.StringVar(&readBufferSize, "read-buffer-size", winningDefault(utils.GetEnvVariable("READ_BUFFER_SIZE"), DEFAULT_READ_BUFFER_SIZE), "Websocket read buffer size")
	flag.StringVar(&writeBufferSize, "write-buffer-size", winningDefault(utils.GetEnvVariable("WRITE_BUFFER_SIZE"), DEFAULT_WRITE_BUFFER_SIZE), "Websocket write buffer size")
	flag.Parse()

	return &ServerOptions{
		Env:             env,
		Domain:          domain,
		Port:            toInt(port),
		LogLevel:        handleLogging(loglevel),
		ReadBufferSize:  toInt(readBufferSize),
		WriteBufferSize: toInt(writeBufferSize),
	}
}
