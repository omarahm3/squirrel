package client

import (
	"flag"
	"fmt"
	"os"

	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap/zapcore"
)

type ClientOptions struct {
	Env          string
	Domain       *common.Domain
	LogLevel     zapcore.Level
	PeerId       string
	Listen       bool
	Output       bool
	UrlClipboard bool
}

const (
	DEFAULT_ENVIRONMENT = "prod"
	DEFAULT_DOMAIN      = "localhost:3000"
	DEFAULT_LOG_LEVEL   = "error"
)

var (
	env          string
	domain       string
	loglevel     string
	peer         string
	listen       bool
	output       bool
	urlClipboard bool
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

	flag.StringVar(&env, "env", common.WinningDefault(common.GetEnvVariable("APP_ENV"), env, DEFAULT_ENVIRONMENT), "Client environment (prod|dev)")
	flag.StringVar(&domain, "domain", common.WinningDefault(common.GetEnvVariable("DOMAIN"), domain, DEFAULT_DOMAIN), "Server domain")
	flag.StringVar(&loglevel, "log", common.WinningDefault(common.GetEnvVariable("LOG_LEVEL"), loglevel, DEFAULT_LOG_LEVEL), "Log level")
	flag.StringVar(&peer, "peer", "", "Peer client ID")
	flag.BoolVar(&listen, "listen", false, "Initiate in listen mode to listen to peer")
	flag.BoolVar(&listen, "l", false, "Initiate in listen mode to listen to peer")
	flag.BoolVar(&output, "show-output", false, "Print output stream to stdout")
	flag.BoolVar(&output, "o", false, "Print output stream to stdout")
	flag.BoolVar(&urlClipboard, "copy-url", false, "Copy shareable link to clipboard")
	flag.BoolVar(&urlClipboard, "u", false, "Copy shareable link to clipboard")
	flag.Parse()

	return &ClientOptions{
		Env:          env,
		Domain:       common.BuildDomain(domain, env),
		LogLevel:     common.GetLogLevelFromString(loglevel),
		PeerId:       peer,
		Listen:       listen,
		Output:       output,
		UrlClipboard: urlClipboard,
	}
}
