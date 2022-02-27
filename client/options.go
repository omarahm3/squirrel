package client

import (
	"flag"
	"fmt"
	"os"

	"github.com/omarahm3/squirrel/utils"
	"go.uber.org/zap/zapcore"
)

type ClientOptions struct {
	Env      string
	Domain   string
	LogLevel zapcore.Level
}

const (
	DEFAULT_ENVIRONMENT = "prod"
	DEFAULT_DOMAIN      = "localhost:3000"
	DEFAULT_LOG_LEVEL   = "error"
)

func fprintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func InitOptions() *ClientOptions {
	flag.Usage = func() {
		fprintf("Usage of %s:\n", os.Args[0])
		fprintf(" %s [options]\n", os.Args[0])
		fprintf("Options:\n")
		flag.PrintDefaults()
	}

	var (
		env      string = utils.GetEnvVariable("APP_ENV")
		domain   string = utils.GetEnvVariable("DOMAIN")
		loglevel string = utils.GetEnvVariable("LOG_LEVEL")
	)

	flag.StringVar(&env, "env", utils.WinningDefault(utils.GetEnvVariable("APP_ENV"), DEFAULT_ENVIRONMENT), "Client environment (prod|dev)")
	flag.StringVar(&domain, "domain", utils.WinningDefault(utils.GetEnvVariable("DOMAIN"), DEFAULT_DOMAIN), "Server domain")
	flag.StringVar(&loglevel, "log", utils.WinningDefault(utils.GetEnvVariable("LOG_LEVEL"), DEFAULT_LOG_LEVEL), "Log level")
	flag.Parse()

	return &ClientOptions{
		Env:      env,
		Domain:   domain,
		LogLevel: utils.GetLogLevelFromString(loglevel),
	}
}
