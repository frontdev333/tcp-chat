package hub

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"frontdev333/tcp-chat/internal"
	"frontdev333/tcp-chat/internal/chat"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

type Hub struct {
	clients    map[*internal.Client]bool
	broadcast  chan chat.ChatMessage
	register   chan *internal.Client
	unregister chan *internal.Client
	requests   chan Request
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*internal.Client]bool),
		broadcast:  make(chan chat.ChatMessage),
		register:   make(chan *internal.Client),
		unregister: make(chan *internal.Client),
		requests:   make(chan Request, 1),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			h.cleanupClient(client)
		case request := <-h.requests:
			request.execute(h.clients)
		case msg := <-h.broadcast:
			for _, c := range h.BroadcastMessage(msg) {
				h.cleanupClient(c)
			}
		}
	}
}

func (h *Hub) BroadcastMessage(msg chat.ChatMessage) []*internal.Client {
	var failed []*internal.Client

	for client := range h.clients {
		if client.ID == msg.PureClientID() {
			continue
		}

		if _, err := client.Conn.Write([]byte(chat.FormatMessage(msg))); err != nil {
			slog.Error(err.Error())
			failed = append(failed, client)
			continue
		}
	}
	return failed
}

func (h *Hub) RegisterClient(ctx context.Context, conn net.Conn, history *chat.History) {
	client := h.setupClientConnection(ctx, conn)
	h.register <- client
	slog.Info("Client connected to server\n", "UserID", client.ID, "Total", h.GetClientCount())
	h.SendMessagesHistory(client, history)
	go h.handleClientMessages(ctx, client, history)
}

func (h *Hub) handleClientMessages(ctx context.Context, client *internal.Client, history *chat.History) {
	defer func() {
		h.unregister <- client
		slog.Info("Client disconnected from server\n", "userID", client.ID, "Total", h.GetClientCount())
	}()

	scanner := bufio.NewScanner(client.Conn)
	for {
		client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
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
		rawText := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(rawText, "/") {
			h.HandleCommand(client, strings.TrimSpace(rawText), history)
			continue
		}

		msg := chat.ParseIncomingMessage(rawText, client.ID)
		slog.Info("got client message", "userID", client.ID, "message", rawText)
		h.broadcast <- msg
		history.Add(&msg)
	}
}

func (h *Hub) GetActiveClients() []string {
	resChan := make(chan ActiveClientsResult, 1)
	h.requests <- &ActiveClientsRequest{resChan}

	return <-resChan
}

func (h *Hub) GetClientCount() int {
	resChan := make(chan ClientsCountResult, 1)
	h.requests <- &ClientsCountRequest{resChan}
	return int(<-resChan)
}

func (h *Hub) setupClientConnection(ctx context.Context, conn net.Conn) *internal.Client {
	client := &internal.Client{
		ID:       internal.GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
	}

	client.Conn.SetDeadline(time.Now().Add(time.Second * 30))

	return client
}

func (h *Hub) cleanupClient(client *internal.Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		client.Conn.Close()
	}
	msg := fmt.Sprintf("Client %s disconnected\n", client.ID)
	h.broadcast <- chat.ParseIncomingMessage(msg, client.ID)
}

func (h *Hub) SendUserList(client *internal.Client) {
	var res bytes.Buffer
	if _, err := fmt.Fprintf(&res, "Online users (%d):", h.GetClientCount()); err != nil {
		slog.Error(err.Error())
		return
	}

	usersString := strings.Join(h.GetActiveClients(), ", ") + "\n"
	if _, err := client.Conn.Write([]byte(usersString)); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Hub) SendMessagesHistory(client *internal.Client, history *chat.History) {
	var res bytes.Buffer

	for _, msg := range history.GetLastMessages() {
		res.WriteString(chat.FormatMessage(*msg) + "\n")
	}

	if _, err := client.Conn.Write(res.Bytes()); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Hub) HandleCommand(client *internal.Client, command string, history *chat.History) {
	switch command {
	case "/users":
		h.SendUserList(client)
	case "/history":
		h.SendMessagesHistory(client, history)
	case "/help":
		h.HandleHelpCommand(client)
	case "/quit":
		os.Exit(1)
	default:
		if _, err := client.Conn.Write([]byte("unknown command\n")); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (h *Hub) HandleHelpCommand(client *internal.Client) {
	helpMsg := `Available commands:
  /help  - Show this help message
  /users - List all online users
  /history - Show recent chat history
  /quit  - Leave the chat room
`
	if _, err := client.Conn.Write([]byte(helpMsg)); err != nil {
		slog.Error(err.Error())
	}
}
