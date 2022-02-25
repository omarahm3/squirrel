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

func InitLogging() {
	var config zap.Config

	if GetEnv() == "dev" {
		config = zap.NewDevelopmentConfig()
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level.SetLevel(zap.ErrorLevel)
	}

	config.OutputPaths = []string{
		fmt.Sprintf("%s/.squirrely.log", GetEnvVariable("HOME")),
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
