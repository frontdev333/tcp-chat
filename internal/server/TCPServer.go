package server

import (
	"context"
	"errors"
	"frontdev333/tcp-chat/internal/chat"
	"frontdev333/tcp-chat/internal/hub"
	"log/slog"
	"net"
)

func StartEchoServer(
	ctx context.Context,
	hub *hub.Hub,
	history *chat.History,
	logger *slog.Logger,
	port string,
) error {
	logger.Info("server started")
	listener, err := net.Listen("tcp", port)
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
			go hub.RegisterClient(ctx, conn, history)
		}
	}
}
