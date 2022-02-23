package main

import "log"

// Maintain the set of active clients
type Hub struct {
	clients    map[string]*Client
	broadcast  chan []byte
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
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
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
		delete(hub.clients, clientId)
		if removeChannel {
			close(client.send)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Println("Registering client:", client.id)
			h.clients[client.id] = client
		case info := <-h.update:
			log.Printf("Updating client ID: [%s] to [%s]", info.id, info.client.id)
			removeClient(h, info.id, false)
			h.clients[info.client.id] = info.client

		case client := <-h.unregister:
			removeClient(client.hub, client.id, true)
		case message := <-h.broadcast:
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client.id)
					close(client.send)
				}
			}
		}
	}
}
