package internal

import (
	"bufio"
	"context"
	"log/slog"
	"net"
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
			h.unregister <- client
			continue
		}
	}
}

func (h *Hub) RegisterClient(ctx context.Context, client *Client) {
	h.register <- client
	go h.handleClientMessages(ctx, client)
	slog.Info("Client connected to server\n", "UserID", client.ID)
}

func (h *Hub) handleClientMessages(ctx context.Context, client *Client) {
	defer func() {
		h.unregister <- client
		slog.Info("Client disconnected from server\n", "userID", client.ID)
	}()

	scanner := bufio.NewScanner(client.Conn)
	for {
		if !scanner.Scan() {
			err := scanner.Err()
			if err == nil {
				return
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if ctx.Err() != nil {
					slog.Error(ctx.Err().Error())
					return
				}
				continue
			}

			slog.Error(err.Error())
			return
		}
		rawText := scanner.Text()
		msg := ParseIncomingMessage(rawText, client.ID)

		slog.Info("got client message", "userID", client.ID, "message", rawText)
		h.broadcast <- msg
	}

}
