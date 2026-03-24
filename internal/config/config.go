package config

import (
	"errors"
	"flag"
	"log/slog"
)

type ServerConfig struct {
	Port               string
	LogLevel           slog.Level
	MessageHistorySize int
	MaxConnections     int
}

func ParseCommandLineArgs() (ServerConfig, error) {
	portCmd := flag.String("port", "8080", "port")
	logLvlCmd := flag.String("log-level", "INFO", "logging level")
	msgHistoryCmd := flag.Int("history-size", 64, "messages history size, power of two")
	maxConnCmd := flag.Int("max-connections", 100, "number of maximum connections")
	flag.Parse()

	if err := validateSrvrCfgData(
		*portCmd,
		*logLvlCmd,
		*msgHistoryCmd,
		*maxConnCmd,
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
		LogLevel:           level,
		MessageHistorySize: *msgHistoryCmd,
		MaxConnections:     *maxConnCmd,
	}, nil
}

func validateSrvrCfgData(
	port string,
	level string,
	historySize int,
	maxConn int,
) error {

	if port == "" {
		return errors.New("port value can not be empty")
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

	if maxConn < 1 {
		return errors.New("maximum connections number can't be less then 1")
	}

	return nil
}
