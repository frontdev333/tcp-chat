package main

import (
	"context"
	"frontdev333/tcp-chat/internal"
	"frontdev333/tcp-chat/internal/app"
	"frontdev333/tcp-chat/internal/config"
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

	logger := slog.New(internal.NewHandler(os.Stdout, "TCP-CHAT", config.LogLevel))

	app.Run(ctx, logger, config)
}
