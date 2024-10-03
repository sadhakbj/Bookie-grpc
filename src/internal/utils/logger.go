package utils

import (
	"log/slog"
	"os"
	"time"
)

func InitializeLogger(name string, addSource bool) *slog.Logger {
	handler := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
				a.Value = slog.Int64Value(time.Now().Unix())
			}
			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handler).WithAttrs([]slog.Attr{
		slog.String("service", name),
	}))
	return logger
}
