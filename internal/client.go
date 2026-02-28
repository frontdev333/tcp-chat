package internal

import (
	"bufio"
	"context"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID       string    `json:"id"`
	Conn     net.Conn  `json:"conn"`
	JoinTime time.Time `json:"join_time"`
}

func (c *Client) HandleClient(ctx context.Context) error {
	defer func() {
		slog.Info("Client disconnected from server\n", "userID", c.ID)
	}()

	slog.Info("Client connected to server\n", "UserID", c.ID)
	scanner := bufio.NewScanner(c.Conn)
	for {
		if !scanner.Scan() {
			err := scanner.Err()
			if err != nil {
				slog.Error(err.Error())
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}

			return err
		}
		rawText := scanner.Text()
		msg := ParseIncomingMessage(rawText, c.ID)

		slog.Info("got client message", "userID", c.ID, "message", rawText)
		if _, err := c.Conn.Write([]byte(FormatMessage(msg) + "\n")); err != nil {
			slog.Warn(err.Error())
			continue
		}
	}
}

func GenerateClientID() string {
	clientIDUint := uint64(uuid.New().ID())
	return strconv.FormatUint(clientIDUint, 10)
}
