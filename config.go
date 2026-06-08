package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func initLogging() (*slog.Logger, func()) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	logDir := filepath.Join(home, ".askd")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("failed to create log directory %s: %v", logDir, err)
		return nil, func() {}
	}

	logPath := filepath.Join(logDir, "app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed to open log file %s: %v", logPath, err)
		return nil, func() {}
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	return logger, func() { f.Close() }
}
