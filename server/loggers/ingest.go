package loggers

import (
	"log/slog"
	"os"
	"path/filepath"
)

var ingestLogger *slog.Logger

func initializeIngestLogger() error {
	if ingestLogger != nil {
		return nil
	}
	logPath := filepath.Join("logs", "ingest.log")
	// remove on every restart for now
	_ = os.Remove(logPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	ingestLogger = slog.New(handler)
	return nil
}

func IngestLogger() *slog.Logger {
	return ingestLogger
}
