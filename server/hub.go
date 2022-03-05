package server

import (
	"go.uber.org/zap"
)

// Maintain the set of active clients
type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan struct {
		message  []byte
		clientId string
	}
	send chan struct {
		message  []byte
		clientId string
	}
	update chan struct {
		id     string
		client *Client
	}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast: make(chan struct {
			message  []byte
			clientId string
		}),
		send: make(chan struct {
			message  []byte
			clientId string
		}),
		update: make(chan struct {
			id     string
			client *Client
		}),
	}
}

// removeChannel is a flag must be set to true to remove client send channel
// This flag is added to avoid removing client channel when updating the client
func (h *Hub) RemoveClient(clientId string, removeChannel bool) {
	if client, ok := hub.clients[clientId]; ok {
		zap.S().Infow("Removing client",
			"clientId", clientId,
			"removeChannel", removeChannel)

		delete(hub.clients, clientId)
		if removeChannel {
			close(client.send)
		}
	}
}

func (h *Hub) RemoveActiveSubscribers(clientId string) {
	for _, c := range h.clients {
		if c.IsActiveSubscriber() && c.peerId == clientId {
			h.RemoveClient(c.id, true)
		}
	}
}

func (h *Hub) Run() {
	zap.S().Debug("Created clients hub")

	for {
		select {
		case client := <-h.register:
			zap.S().Infow("Adding client to hub",
				"id", client.id,
				"active", client.active,
				"broadcaster", client.broadcaster,
				"subscriber", client.subscriber,
				"peerId", client.peerId)

			h.clients[client.id] = client

		case info := <-h.update:
			zap.S().Infow("Updating client",
				"id", info.id,
				"newId", info.client.id)

			h.RemoveClient(info.id, false)
			h.clients[info.client.id] = info.client

		case client := <-h.unregister:
			zap.S().Infow("Unregistering client",
				"id", client.id)

			// In case broadcaster is disconnecting, then disconnect subscribers too
			if client.IsActiveBroadcaster() {
				h.RemoveActiveSubscribers(client.id)
			}
			h.RemoveClient(client.id, true)

		case message := <-h.broadcast:
			zap.S().Infow("Broadcasting message to peer",
				"clientId", message.clientId)

			for _, client := range h.clients {
				if client.IsActiveSubscriber() && client.peerId == message.clientId {
					zap.S().Debugw("Sending message to client",
						"clientId", client.id,
						"broadcaster", client.broadcaster,
						"subscriber", client.subscriber,
						"active", client.active)

					client.send <- message.message
				}
			}

		case message := <-h.send:
			zap.S().Infow("Sending message to peer",
				"clientId", message.clientId)

			client, ok := h.clients[message.clientId]

			if ok {
				client.send <- message.message
			}

			zap.L().Error("Couldn't find client", zap.String("clientId", message.clientId))
		}
	}
}
