package utils

import (
	"context"
	"log/slog"
)

type slogFanout struct {
	handlers []slog.Handler
}

func newSlogFanout(handlers ...slog.Handler) slog.Handler {
	return slogFanout{handlers: handlers}
}

func (f slogFanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range f.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f slogFanout) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range f.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (f slogFanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for i, handler := range f.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return slogFanout{handlers: handlers}
}

func (f slogFanout) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for i, handler := range f.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return slogFanout{handlers: handlers}
}
