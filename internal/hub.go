package internal

import (
	"log/slog"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan ChatMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close()
			}
		case msg := <-h.broadcast:
			h.BroadcastMessage(msg)
		}
	}
}

func (h *Hub) BroadcastMessage(msg ChatMessage) {
	for client := range h.clients {
		if client.ID == msg.PureClientID() {
			continue
		}

		if _, err := client.Conn.Write([]byte(FormatMessage(msg))); err != nil {
			slog.Error(err.Error())
			return
		}
	}
}
