package server

import (
	"bufio"
	"context"
	"errors"
	"frontdev333/tcp-chat/internal/chat"
	"frontdev333/tcp-chat/internal/config"
	"frontdev333/tcp-chat/internal/hub"
	"log/slog"
	"net"
	"regexp"
	"strings"
)

var validationRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

func StartEchoServer(
	ctx context.Context,
	hub *hub.Hub,
	history *chat.History,
	logger *slog.Logger,
	config config.ServerConfig,
) error {
	logger.Info("server started")
	listener, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			slog.Error(err.Error())
			continue
		}
		select {
		case <-ctx.Done():
			conn.Close()
			return ctx.Err()
		default:
			if hub.GetClientCount() >= config.MaxConnections {
				conn.Close()
				logger.Warn("Maximum connections limit")
				continue
			}
			go performHandshake(ctx, hub, conn, history)
		}
	}

}

func performHandshake(ctx context.Context, h *hub.Hub, conn net.Conn, history *chat.History) {
	if _, err := conn.Write([]byte("Input your nickname. It'll use as your ID: ")); err != nil {
		slog.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	var nick string
	i := 0
	isSuccess := false

out:
	for scanner.Scan() {

		nick = strings.TrimSpace(scanner.Text())

		if !validationRegexp.MatchString(nick) {

			i++
			if i == 3 {
				if _, err := conn.Write([]byte("Your three tries have been spend")); err != nil {
					slog.Error(err.Error())
				}
				conn.Close()
				return
			}
			if _, err := conn.Write([]byte("Allowed chars: alphabet, numbers, _, -\nAnd your nickname must be longer than 3, but shorter than 20\n")); err != nil {
				slog.Error(err.Error())
			}
			continue
		}

		response := make(chan hub.ActiveClientsIDsResult)
		h.Requests <- &hub.ActiveClientsIDsRequest{Response: response}

		for _, id := range <-response {
			if id == nick {
				i++
				if i == 3 {
					if _, err := conn.Write([]byte("Your three tries have been spend")); err != nil {
						slog.Error(err.Error())
					}
					conn.Close()
					return
				}

				if _, err := conn.Write([]byte("This nickname is already is use")); err != nil {
					slog.Error(err.Error())
					continue out
				}
			}
		}
		isSuccess = true
		break
	}

	if !isSuccess {
		conn.Close()
		return
	}

	h.RegisterClient(ctx, conn, nick, history)
}
