package server

import (
	"go.uber.org/zap"
)

// Maintain the set of active clients
type Hub struct {
	clients   map[string]*Client
	broadcast chan struct {
		message  []byte
		clientId string
	}
	register   chan *Client
	unregister chan *Client
	update     chan struct {
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
		update: make(chan struct {
			id     string
			client *Client
		}),
	}
}

// removeChannel is a flag must be set to true to remove client send channel
// This flag is added to avoid removing client channel when updating the client
func removeClient(hub *Hub, clientId string, removeChannel bool) {
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

			removeClient(h, info.id, false)
			h.clients[info.client.id] = info.client

		case client := <-h.unregister:
			zap.S().Infow("Unregistering client",
				"id", client.id)

			// In case broadcaster is disconnecting, then disconnect subscribers too
			if client.IsActiveBroadcaster() {
				for _, c := range h.clients {
					if c.IsActiveSubscriber() && c.peerId == client.id {
						removeClient(c.hub, c.id, true)
						// c.connection.WriteMessage(websocket.CloseMessage, []byte{})
						// c.connection.Close()
					}
				}
			}
			removeClient(client.hub, client.id, true)

		case message := <-h.broadcast:
			zap.S().Infow("Broadcasting message to peer",
				"clientId", message.clientId)

			for _, client := range h.clients {
				// Ignore any client and only accept client that has the link
				if client.broadcaster || !client.subscriber || !client.active || client.peerId != message.clientId {
					zap.S().Debugw("Ignoring broadcasting message to this client",
						"clientId", client.id,
						"broadcaster", client.broadcaster,
						"subscriber", client.subscriber,
						"active", client.active)

					continue
				}

				zap.S().Debugw("Sending message to client",
					"clientId", client.id,
					"broadcaster", client.broadcaster,
					"subscriber", client.subscriber,
					"active", client.active)

				select {
				case client.send <- message.message:
				default:
					delete(h.clients, client.id)
					close(client.send)
				}
			}
		}
	}
}
