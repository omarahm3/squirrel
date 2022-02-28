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
	Domain   *utils.Domain
	LogLevel zapcore.Level
	PeerId   string
	Listen   bool
}

const (
	DEFAULT_ENVIRONMENT = "prod"
	DEFAULT_DOMAIN      = "localhost:3000"
	DEFAULT_LOG_LEVEL   = "error"
)

var (
	env      string
	domain   string
	loglevel string
	peer     string
	listen   bool
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

	flag.StringVar(&env, "env", utils.WinningDefault(utils.GetEnvVariable("APP_ENV"), env, DEFAULT_ENVIRONMENT), "Client environment (prod|dev)")
	flag.StringVar(&domain, "domain", utils.WinningDefault(utils.GetEnvVariable("DOMAIN"), domain, DEFAULT_DOMAIN), "Server domain")
	flag.StringVar(&loglevel, "log", utils.WinningDefault(utils.GetEnvVariable("LOG_LEVEL"), loglevel, DEFAULT_LOG_LEVEL), "Log level")
	flag.StringVar(&peer, "peer", "", "Peer client ID")
	flag.BoolVar(&listen, "listen", false, "Initiate in listen mode to listen to peer")
	flag.Parse()

	return &ClientOptions{
		Env:      env,
		Domain:   utils.BuildDomain(domain, env),
		LogLevel: utils.GetLogLevelFromString(loglevel),
		PeerId:   peer,
		Listen:   listen,
	}
}
