package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/al002/zbittorrent/internal/config"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *config.LogConfig) (*Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Log config is nil")
	}

	level := parseLevel(cfg.Level)

	writer, err := getWriter(cfg.Dir)

	if err != nil {
		return nil, err
	}

	handler := createHandler(writer, level, cfg.Format)

	logger := &Logger{
		Logger: slog.New(handler),
	}

	slog.SetDefault(logger.Logger)

	return logger, nil

}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getWriter(dir string) (*os.File, error) {
	if dir == "" {
		return os.Stdout, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("Failed to create log directory: %v", err)
	}

	logPath := filepath.Join(dir, "zbittorrent.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Failed to open log file: %v", err)
	}

	return f, nil
}

func createHandler(writer *os.File, level slog.Level, format string) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	jsonHandler := slog.NewJSONHandler(writer, opts)

	switch strings.ToLower(format) {
	case "json":
		return jsonHandler
	case "text":
		return slog.NewTextHandler(writer, opts)
	default:
		return jsonHandler
	}
}

func (l *Logger) Close() error {
	if h, ok := l.Handler().(interface{ Close() error }); ok {
		return h.Close()
	}

	return nil
}
