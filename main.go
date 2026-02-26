package main

import (
	"context"
	"frontdev333/tcp-chat/internal"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer cancel()

	go func() {
		err := internal.StartEchoServer(ctx, ":8080")
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}()

	<-ctx.Done()
	slog.Warn("shutdown started")

}
