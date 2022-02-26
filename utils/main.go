package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
