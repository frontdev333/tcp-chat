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

	config, err := internal.ParseCommandLineArgs()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	logger := slog.New(internal.NewHandler(os.Stdout, "TCP-CHAT", config.LogLevel))
	hub := hub.NewHub(logger, time.Now())
	history := chat.NewHistory(config.MessageHistorySize)

	internal.PrintStartupBanner(config)

	go func() {
		if err = server.StartHTTPMonitoring(hub, "8081"); err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}()
	go hub.Run()

	go func() {
		err := server.StartEchoServer(ctx, hub, history, logger, &config, ":"+config.Port)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}()

	<-ctx.Done()
	logger.Warn("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := hub.Shutdown(shutdownCtx); err != nil {
		slog.Error(err.Error())
	}

	logger.Info("server successfully stopped")

}
