package utils

import (
	"log"
	"os"

	"github.com/google/uuid"
)

func GenerateUUID() string {
  id, err := uuid.NewRandom()

  if err != nil {
    log.Fatalf("Error creating random UUID for this client: %+v", err)
    os.Exit(1)
  }

  return id.String()
}
