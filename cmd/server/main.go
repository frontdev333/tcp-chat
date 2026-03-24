package main

import (
	"context"
	"frontdev333/tcp-chat/internal/app"
	"frontdev333/tcp-chat/internal/config"
	"frontdev333/tcp-chat/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	config, err := config.ParseCommandLineArgs()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	logger := slog.New(logger.NewHandler(os.Stdout, "TCP-CHAT", config.LogLevel))

	app.Run(ctx, logger, config)
}
