// Package utils provides a slog handler that enriches log records with
// the trace_id and span_id of the span recorded on the log context.
package utils

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type traceHandler struct {
	slog.Handler
}

// NewTraceLoggerHandler wraps the given handler so every record carries the
// current span's trace_id and span_id when a span is present in the context.
func NewTraceLoggerHandler(h slog.Handler) slog.Handler {
	return &traceHandler{Handler: h}
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() && sc.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()))
		if sc.HasSpanID() {
			r.AddAttrs(slog.String("span_id", sc.SpanID().String()))
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithGroup(name)}
}
