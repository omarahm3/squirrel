package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/omarahm3/squirrel/internal/pkg/common"
)

func TestInitClient(t *testing.T) {
	defer ResetTesting(nil)
	options = InitOptions()

	server := newServer()
	defer server.Close()

	overrideClientOptionsDomain(server)

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

func TestClientSubscriberAckMessage(t *testing.T) {
	events = make(chan string)

	defer ResetTesting(nil)
	options = InitOptions()

	server := newServer()
	defer server.Close()

	overrideClientOptionsDomain(server)

	connection := InitClient()
	defer connection.Close()

	jsonMessage := common.Message{
		Id:    clientId,
		Event: EVENT_SUBSCRIBER_ACK,
		Payload: common.SubscriberConnectedMessage{
			Connected: true,
		},
	}

	connection.WriteJSON(jsonMessage)

	go HandleIncomingMessages(connection)

	// Subscriber must be connected so that we receive message on events channel
	event := <-events

	if event != jsonMessage.Event {
		t.Errorf("expected %q, actual %q", jsonMessage.Event, event)
	}
}

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

func overrideClientOptionsDomain(server *httptest.Server) {
	url := strings.TrimPrefix(server.URL, "http://")

	// override the global options and use server URL
	options.Domain = common.BuildDomain(url, "dev")
}

func newServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(echo))
}