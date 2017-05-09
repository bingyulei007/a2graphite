package main

import (
	"fmt"
	"github.com/op/go-logging"
	"os"
	"strings"
)

func setupLogging(config *logConfig) {
	// NOTE the correct order is:
	// 1. SetFormatter
	// 2. SetBackend
	// 3. SetLevel
	// SetLevel() sets level to logging.defaultBackend, so must be called after SetBackend().
	// A backend fetches formatter on it's first use, it's better to call SetFormatter() before SetBackend().

	format := logging.MustStringFormatter(`[%{time:15:04:05.000}][%{module}][%{level}] %{message}`)
	logging.SetFormatter(format)

	logFile, err := os.OpenFile(config.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to open logfile:", err)
		os.Exit(1)
	}
	backend := logging.NewLogBackend(logFile, "", 0)
	logging.SetBackend(backend)

	var level logging.Level
	switch strings.ToLower(config.Level) {
	case "critical":
		level = logging.CRITICAL
	case "error":
		level = logging.ERROR
	case "warning":
		level = logging.WARNING
	case "notice":
		level = logging.NOTICE
	case "info":
		level = logging.INFO
	case "debug":
		level = logging.DEBUG
	default:
		fmt.Println("Invalid log level:", config.Level)
		os.Exit(1)
	}
	logging.SetLevel(level, "")
}
