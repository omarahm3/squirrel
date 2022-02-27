package client

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/omarahm3/squirrel/utils"
	"go.uber.org/zap/zapcore"
)

type Domain struct {
	Public    string
	Websocket string
}

type ClientOptions struct {
	Env      string
	Domain   *Domain
	LogLevel zapcore.Level
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
)

func fprintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func buildDomain(domain string, env string) *Domain {
	var s string

	if env == "prod" {
		s = "s"
	}

	public := url.URL{
		Scheme: fmt.Sprintf("http%s", s),
		Host:   domain,
	}

	websocket := url.URL{
		Scheme: fmt.Sprintf("ws%s", s),
		Host:   domain,
	}

	return &Domain{
		Public:    public.String(),
		Websocket: websocket.String(),
	}
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
	flag.Parse()

	return &ClientOptions{
		Env:      env,
		Domain:   buildDomain(domain, env),
		LogLevel: utils.GetLogLevelFromString(loglevel),
	}
}
