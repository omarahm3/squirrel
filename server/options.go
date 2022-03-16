package server

import (
	"flag"
	"fmt"
	"os"

	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap/zapcore"
)

type ServerOptions struct {
	Env             string
	Domain          *common.Domain
	Port            int
	LogLevel        zapcore.Level
	ReadBufferSize  int
	WriteBufferSize int
	MaxMessageSize  int64
}

const (
	DEFAULT_ENVIRONMENT       = "prod"
	DEFAULT_DOMAIN            = "localhost:3000"
	DEFAULT_PORT              = "3000"
	DEFAULT_LOG_LEVEL         = "warn"
	DEFAULT_READ_BUFFER_SIZE  = "0"
	DEFAULT_WRITE_BUFFER_SIZE = "0"
	DEFAULT_MAX_MESSAGE_SIZE  = "1024"
)

var (
	env             string
	domain          string
	port            string
	loglevel        string
	readBufferSize  string
	writeBufferSize string
	maxMessageSize  string
)

func fprintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func InitOptions() *ServerOptions {
	flag.Usage = func() {
		fprintf("Usage of %s:\n", os.Args[0])
		fprintf(" %s [options]\n", os.Args[0])
		fprintf("Options:\n")
		flag.PrintDefaults()
		fprintf("%s configures server run.\n", os.Args[0])
	}

	flag.StringVar(&env, "env", common.WinningDefault(common.GetEnvVariable("APP_ENV"), env, DEFAULT_ENVIRONMENT), "Server environment (prod|dev)")
	flag.StringVar(&domain, "domain", common.WinningDefault(common.GetEnvVariable("DOMAIN"), domain, DEFAULT_DOMAIN), "Server domain")
	flag.StringVar(&loglevel, "log", common.WinningDefault(common.GetEnvVariable("LOG_LEVEL"), loglevel, DEFAULT_LOG_LEVEL), "Log level")
	flag.StringVar(&port, "port", common.WinningDefault(common.GetEnvVariable("PORT"), port, DEFAULT_PORT), "Server port")
	flag.StringVar(&readBufferSize, "read-buffer-size", common.WinningDefault(common.GetEnvVariable("READ_BUFFER_SIZE"), readBufferSize, DEFAULT_READ_BUFFER_SIZE), "Websocket read buffer size")
	flag.StringVar(&writeBufferSize, "write-buffer-size", common.WinningDefault(common.GetEnvVariable("WRITE_BUFFER_SIZE"), writeBufferSize, DEFAULT_WRITE_BUFFER_SIZE), "Websocket write buffer size")
	flag.StringVar(&maxMessageSize, "max-message-size", common.WinningDefault(common.GetEnvVariable("MAX_MESSAGE_SIZE"), maxMessageSize, DEFAULT_MAX_MESSAGE_SIZE), "Websocket maximum message size")
	flag.Parse()

	return &ServerOptions{
		Env:             env,
		Domain:          common.BuildDomain(domain, env),
		Port:            common.StrToInt(port),
		LogLevel:        common.GetLogLevelFromString(loglevel),
		ReadBufferSize:  common.StrToInt(readBufferSize),
		WriteBufferSize: common.StrToInt(writeBufferSize),
		MaxMessageSize:  common.StrToInt64(maxMessageSize),
	}
}
