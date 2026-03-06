package main

import (
	"context"
	"frontdev333/tcp-chat/internal/hub"
	"frontdev333/tcp-chat/internal/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer cancel()

	hub := hub.NewHub()
	go hub.Run()

	go func() {
		err := server.StartEchoServer(ctx, hub, ":8080")
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}()

	<-ctx.Done()
	slog.Warn("shutdown started")

}
