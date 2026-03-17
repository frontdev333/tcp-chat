package internal

import (
	"errors"
	"flag"
	"log/slog"
)

type ServerConfig struct {
	Port               string
	MaxConnections     int
	LogLevel           slog.Level
	MessageHistorySize int
}

func ParseCommandLineArgs() (ServerConfig, error) {
	portCmd := flag.String("port", "8080", "port")
	maxConnCmd := flag.Int("max-connections", 100, "maximum connections")
	logLvlCmd := flag.String("log-level", "INFO", "logging level")
	msgHistoryCmd := flag.Int("history-size", 64, "messages history size, power of two")

	flag.Parse()

	if err := validateSrvrCfgData(
		*portCmd,
		*maxConnCmd,
		*logLvlCmd,
		*msgHistoryCmd,
	); err != nil {
		return ServerConfig{}, err
	}

	level := slog.LevelInfo

	switch *logLvlCmd {
	case "DEBUG":
		level = slog.LevelDebug
	case "ERROR":
		level = slog.LevelError
	}

	return ServerConfig{
		Port:               *portCmd,
		MaxConnections:     *maxConnCmd,
		LogLevel:           level,
		MessageHistorySize: *msgHistoryCmd,
	}, nil
}

func validateSrvrCfgData(
	port string,
	maxConn int,
	level string,
	historySize int,
) error {

	if port == "" {
		return errors.New("port value can not be empty")
	}

	if maxConn <= 0 {
		return errors.New("maximum connections number must be greater than zero")
	}

	if level == "" {
		return errors.New("logging level value can not be empty")
	}

	if level != "INFO" && level != "DEBUG" && level != "ERROR" {
		return errors.New("use proper log level: INFO, DEBUG, ERROR")
	}

	if historySize <= 0 {
		return errors.New("chat messages history size number must be greater then zero")
	}

	if historySize&(historySize-1) != 0 {
		return errors.New("messages history size must be number in power of two")
	}

	return nil
}
