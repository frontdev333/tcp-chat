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
	requests   chan ActiveClientsRequest
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		requests:   make(chan ActiveClientsRequest, 1),
	}
}

type ActiveClientsResult []string

type ActiveClientsRequest struct {
	Response chan ActiveClientsResult
}

type ClientsCountResult []string

type ClientsCountRequest struct {
	Response chan ClientsCountResult
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
		case request := <-h.requests:
			h.sendActiveClients(request)
		case msg := <-h.broadcast:
			h.BroadcastMessage(msg)
		}
	}
}

func (h *Hub) sendActiveClients(request ActiveClientsRequest) {
	res := make([]string, 0, len(h.clients))

	for c := range h.clients {
		res = append(res, c.ID)
	}

	request.Response <- res
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
	slog.Info("Client connected to server\n", "UserID", client.ID, "Total", h.GetClientCount())
	go h.handleClientMessages(ctx, client)
}

func (h *Hub) handleClientMessages(ctx context.Context, client *Client) {
	defer func() {
		h.unregister <- client
		slog.Info("Client disconnected from server\n", "userID", client.ID, "Total", h.GetClientCount())
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

func (h *Hub) GetActiveClients() []string {
	resChan := make(chan ActiveClientsResult, 1)
	h.requests <- ActiveClientsRequest{resChan}

	return <-resChan
}

func (h *Hub) GetClientCount() int {
	return len(h.GetActiveClients())
}
