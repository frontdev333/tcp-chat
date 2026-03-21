package hub

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"frontdev333/tcp-chat/internal/chat"
	"frontdev333/tcp-chat/internal/domain"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Hub struct {
	clients                map[*domain.Client]bool
	broadcast              chan domain.ChatMessage
	register               chan *domain.Client
	unregister             chan *domain.Client
	requests               chan Request
	logger                 *slog.Logger
	totalMessagesProcessed atomic.Int64
	activeConnections      atomic.Int64
	errorCount             atomic.Int64
	lastError              atomic.Pointer[string]
	startTime              time.Time
}

func NewHub(logger *slog.Logger, start time.Time) *Hub {
	return &Hub{
		clients:    make(map[*domain.Client]bool),
		broadcast:  make(chan domain.ChatMessage),
		register:   make(chan *domain.Client),
		unregister: make(chan *domain.Client),
		requests:   make(chan Request, 1),
		logger:     logger,
		startTime:  start,
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
			h.BroadcastMessage(msg)
		}
	}
}

func (h *Hub) BroadcastMessage(msg domain.ChatMessage) {
	h.totalMessagesProcessed.Add(1)

	for client := range h.clients {
		if client.ID == msg.PureClientID() {
			continue
		}
		select {
		case client.WriteCh <- msg:
		default:
			go func() {
				h.unregister <- client
			}()
		}
	}
}

func (h *Hub) RegisterClient(ctx context.Context, conn net.Conn, history *chat.History) {
	client := h.setupClientConnection(conn)
	go h.clientWriter(client)
	h.register <- client
	h.activeConnections.Add(1)
	h.logger.Info("Client connected to server\n", "UserID", client.ID, "Total", h.GetClientCount())
	h.SendMessagesHistory(client, history)
	h.handleClientMessages(ctx, client, history)
}

func (h *Hub) clientWriter(client *domain.Client) {
	for msg := range client.WriteCh {
		if _, err := client.Conn.Write([]byte(domain.FormatMessage(msg))); err != nil {
			h.unregister <- client
			h.logger.Error(err.Error())
			h.addErr2Stats()
			return
		}
	}
}

func (h *Hub) handleClientMessages(ctx context.Context, client *domain.Client, history *chat.History) {
	defer func() {
		if r := recover(); r != nil {
			h.handlePanic(r, client)
		}

		msg := fmt.Sprintf("Client %s disconnected\n", client.ID)
		h.broadcast <- domain.ParseIncomingMessage(msg, client.ID)

		h.unregister <- client
		h.logger.Info("Client disconnected from server\n", "userID", client.ID, "Total", h.GetClientCount())
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
					h.logger.Error(ctx.Err().Error())
					h.addErr2Stats()
					return
				}
				h.addErr2Stats()
				h.logger.Info("timeout disconnect", "client_id", client.ID)
				return
			}

			h.logger.Error(err.Error())
			h.addErr2Stats()
			return
		}
		rawText := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(rawText, "/") {
			shouldQuit := h.HandleCommand(client, strings.TrimSpace(rawText), history)
			if shouldQuit {
				return
			}
			continue
		}

		msg := domain.ParseIncomingMessage(rawText, client.ID)
		h.logger.Info("got client message", "userID", client.ID, "message", rawText)
		h.broadcast <- msg
		history.Add(&msg)
	}
}

func (h *Hub) GetActiveClientsIDs() []string {
	resChan := make(chan ActiveClientsIDsResult, 1)
	h.requests <- &ActiveClientsIDsRequest{resChan}

	return <-resChan
}

func (h *Hub) GetClientCount() int {
	resChan := make(chan ClientsCountResult, 1)
	h.requests <- &ClientsCountRequest{resChan}
	return int(<-resChan)
}

func (h *Hub) setupClientConnection(conn net.Conn) *domain.Client {
	client := &domain.Client{
		ID:       domain.GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
		WriteCh:  make(chan domain.ChatMessage, 100),
	}

	client.Conn.SetDeadline(time.Now().Add(time.Second * 30))

	return client
}

func (h *Hub) cleanupClient(client *domain.Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		client.Conn.Close()
		close(client.WriteCh)
		h.activeConnections.Add(-1)
	}
}

func (h *Hub) SendUserList(client *domain.Client) {
	var res bytes.Buffer
	if _, err := fmt.Fprintf(&res, "Online users (%d):", h.GetClientCount()); err != nil {
		h.logger.Error(err.Error())
		h.addErr2Stats()
		return
	}

	usersString := strings.Join(h.GetActiveClientsIDs(), ", ")
	res.WriteString(usersString)
	h.SendMessage(client, res.String())
}

func (h *Hub) SendMessagesHistory(client *domain.Client, history *chat.History) {
	for _, msg := range history.GetLastMessages() {
		client.WriteCh <- *msg
	}
}

func (h *Hub) HandleCommand(client *domain.Client, command string, history *chat.History) bool {
	switch command {
	case "/users":
		h.SendUserList(client)
	case "/history":
		h.SendMessagesHistory(client, history)
	case "/help":
		h.HandleHelpCommand(client)
	case "/time":
		h.HandleTimeCommand(client)
	case "/quit":
		return true
	default:
		h.SendMessage(client, "unknown command")
	}
	return false
}

func (h *Hub) HandleTimeCommand(client *domain.Client) {
	timeMsg := fmt.Sprintf("Server time: %s", time.Now().Format("2006-01-02 15:04:05"))
	h.SendMessage(client, timeMsg)
}

func (h *Hub) HandleHelpCommand(client *domain.Client) {
	helpMsg := `Available commands:
  /help  - Show this help message
  /users - List all online users
  /history - Show recent chat history
  /time - Show current server time
  /quit  - Leave the chat room
`
	h.SendMessage(client, helpMsg)
}

func (h *Hub) SendMessage(client *domain.Client, text string) {
	client.WriteCh <- domain.ChatMessage{
		ClientID:    uuid.NewString(),
		ClientType:  "system",
		Content:     text,
		MessageID:   uuid.NewString(),
		MessageType: "notification",
		Timestamp:   time.Now(),
	}
}

func (h *Hub) handlePanic(r any, client *domain.Client) {
	h.logger.Error("handled panic", "error", r, "client_id", client.ID)
	h.addErr2Stats()
}

func (h *Hub) addErr2Stats() {
	h.errorCount.Add(1)
	errTime := time.Now().Format("2006/01/02 15:04:05")
	h.lastError.Store(&errTime)
}

func (h *Hub) GetStats() ServerStats {

	uptime := time.Since(h.startTime).Seconds()
	var lstErr string
	if p := h.lastError.Load(); p != nil {
		lstErr = *p
	}

	ttlMsgs := h.totalMessagesProcessed.Load()
	var msgRPMin float64

	if uptime > 0 {
		msgRPMin = float64(ttlMsgs) / uptime * 60
	}

	return ServerStats{
		ActiveConnections:      h.activeConnections.Load(),
		TotalMessagesProcessed: ttlMsgs,
		UptimeSeconds:          uptime,
		ErrorCount:             h.errorCount.Load(),
		LastError:              lstErr,
		ServerStartTime:        h.startTime.String(),
		MessageRatePerMinute:   msgRPMin,
	}
}

func (h *Hub) getActiveClients() []*domain.Client {
	response := make(chan ActiveClientsResult)
	h.requests <- &ActiveClientsRequest{response}
	return <-response
}

func (h *Hub) notifyShutdown(msgText string) {
	h.broadcast <- domain.ChatMessage{
		ClientID:    uuid.NewString(),
		ClientType:  "system",
		Content:     msgText,
		MessageID:   uuid.NewString(),
		MessageType: "notification",
		Timestamp:   time.Now(),
	}
}

func (h *Hub) Shutdown(ctx context.Context) error {
	h.notifyShutdown("Server is shutting down. Goodbye!")
	clients := h.getActiveClients()

	select {
	case <-ctx.Done():
	case <-time.After(2):
	}
	for _, c := range clients {
		c.Conn.Close()
	}

	return nil
}
