package utils

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Domain struct {
	Public    string
	Websocket string
}

type LoggerOptions struct {
	Env         string
	LogLevel    zapcore.Level
	LogFileName string
}

func GetEnv() string {
	env := os.Getenv("APP_ENV")

	if env == "" {
		return "prod"
	}

	return "dev"
}

func GetEnvVariable(variable string) string {
	return os.Getenv(variable)
}

func InitLogging(options LoggerOptions) {
	var config zap.Config

	if options.Env == "dev" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level.SetLevel(options.LogLevel)
	config.OutputPaths = []string{
		fmt.Sprintf("%s/%s", GetEnvVariable("HOME"), options.LogFileName),
		"stdout",
	}

	config.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.UTC().Format("2006-01-02T15:04:05Z0700"))
	})
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.FunctionKey = "Function"
	logger, err := config.Build()

	if err != nil {
		fmt.Println("Error setting up logger:", err)
		os.Exit(1)
	}

	zap.ReplaceGlobals(logger)
}

func GenerateUUID() string {
	id, err := uuid.NewRandom()

	if err != nil {
		log.Fatalf("Error creating random UUID for this client: %+v", err)
		os.Exit(1)
	}

	return id.String()
}

func FatalError(message string, err error) {
	zap.L().Error(message, zap.Error(err))
	os.Exit(1)
}

func StrToInt64(value string) int64 {
	intVal, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		fmt.Println("Error converting value to int64", err)
		os.Exit(1)
	}

	return intVal
}

func StrToInt(value string) int {
	intVal, err := strconv.Atoi(value)

	if err != nil {
		fmt.Println("Error converting value to int", err)
		os.Exit(1)
	}

	return intVal
}

func GetLogLevelFromString(loglevel string) zapcore.Level {
	level, err := zapcore.ParseLevel(loglevel)

	if err != nil {
		fmt.Println("Invalid log level returned, setting default log level, Error: ", err)
		level = zap.ErrorLevel
	}

	return level
}

// Function will pick first argument if it was not empty, or it will loop over the rest of the arguments
// And pick the first not empty one
func WinningDefault(value string, values ...string) string {
	if value == "" {
		for _, alternative := range values {
			if alternative != "" {
				return alternative
			}
		}
	}
	return value
}

func BuildDomain(domain string, env string) *Domain {
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
