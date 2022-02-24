package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func HandleLogLevel() {
  config := zap.NewProductionConfig()
  config.OutputPaths = []string{
    "./log.txt",
    "stdout",
  }
  config.Level.SetLevel(zap.DebugLevel)
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
