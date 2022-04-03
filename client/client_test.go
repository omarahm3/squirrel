package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/internal/pkg/common"
)

func echo(w http.ResponseWriter, r *http.Request) {
  connection, err := websocket.Upgrade(w, r, nil, 0, 0)

  if err != nil {
    return
  }

  defer connection.Close()

  for {
    messageType, message, err := connection.ReadMessage()

    if err != nil {
      break
    }
    
    err = connection.WriteMessage(messageType, message)

    if err != nil {
      break
    }
  }
}

func TestInitClient(t *testing.T) {
  defer ResetTesting(nil)
  options = InitOptions()

  server := httptest.NewServer(http.HandlerFunc(echo))
  defer server.Close()

  url := strings.TrimPrefix(server.URL, "http://")
  
  // override the global options and use server URL
  options.Domain = common.BuildDomain(url, "dev")
  
  connection := InitClient()
  defer connection.Close()
  
  expected := "test"
  
  if err := connection.WriteMessage(websocket.TextMessage, []byte(expected)); err != nil {
    t.Fatal("error sending text message", err)
  }

  _, message, err := connection.ReadMessage()

  if err != nil {
    t.Fatal("error reading message", err)
  }

  if string(message) != expected {
    t.Errorf("expected %q, actual %q", expected, string(message))
  }
}
