// Package logging membuat slog.Logger terstruktur.
// JSON di production (mudah diparse log aggregator), teks di development.
package logging

import (
	"io"
	"log/slog"
	"strings"
)

func New(level string, production bool, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	if production {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
