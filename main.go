package main

import (
	"context"
	"frontdev333/tcp-chat/internal"
	"frontdev333/tcp-chat/internal/chat"
	"frontdev333/tcp-chat/internal/hub"
	"frontdev333/tcp-chat/internal/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer cancel()

	logger := slog.New(internal.NewHandler(os.Stdout, "TCP-CHAT", slog.LevelInfo))
	stats := &internal.ServerStats{}
	hub := hub.NewHub(logger, stats, time.Now())
	history := chat.NewHistory()
	go hub.Run()

	go func() {
		err := server.StartEchoServer(ctx, hub, history, logger, ":8080")
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}()

	<-ctx.Done()
	logger.Warn("shutdown signal received!")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := hub.Shutdown(shutdownCtx); err != nil {
		slog.Error(err.Error())
	}

	logger.Info("server successfully stopped")

}
