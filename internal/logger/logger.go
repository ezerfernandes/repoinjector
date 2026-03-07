package logger

import (
	"log/slog"
	"os"
)

// Init initializes the default slog logger. When verbose is true, the
// minimum level is set to Debug; otherwise it defaults to Warn so that
// only warnings and errors are shown during normal operation.
func Init(verbose bool) {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}
