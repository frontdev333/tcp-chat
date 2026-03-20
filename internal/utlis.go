package internal

import (
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"
)

func PrintStartupBanner(config ServerConfig) {
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
