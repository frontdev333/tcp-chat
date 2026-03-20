package app

import (
	"context"
	"fmt"
	"frontdev333/tcp-chat/internal/chat"
	"frontdev333/tcp-chat/internal/config"
	"frontdev333/tcp-chat/internal/hub"
	"frontdev333/tcp-chat/internal/server"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"
)

func Run(ctx context.Context, logger *slog.Logger, config config.ServerConfig) {
	printStartupBanner(config)

	hub := hub.NewHub(logger, time.Now())
	history := chat.NewHistory(config.MessageHistorySize)
	go func() {
		if err := server.StartHTTPMonitoring(hub, "8081"); err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}()
	go hub.Run()

	go func() {
		err := server.StartEchoServer(ctx, hub, history, logger, ":"+config.Port)
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

func printStartupBanner(config config.ServerConfig) {
	banner := `╔══════════════════════════════════════╗
║         TCP Chat Server              ║
║    Emergency Team Coordination       ║
╚══════════════════════════════════════╝`

	fmt.Println(banner)

	tb := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	level := "INFO"
	switch config.LogLevel {
	case slog.LevelDebug:
		level = "DEBUG"
	case slog.LevelError:
		level = "ERROR"
	}

	fmt.Fprintf(tb, "TCP Server Port:\t%s\n", config.Port)
	fmt.Fprintf(tb, "Log Level:\t%s\n", level)
	tb.Flush()

	fmt.Printf("Connect using: telnet localhost %s\n", config.Port)
}
