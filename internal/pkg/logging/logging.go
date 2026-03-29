package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// New creates a *slog.Logger wrapped with an otelHandler that automatically
// injects trace_id and span_id from the OpenTelemetry span context.
// In dev mode it uses a human-readable text handler; otherwise JSON.
func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	var base slog.Handler
	if env == "dev" {
		base = slog.NewTextHandler(os.Stdout, opts)
	} else {
		base = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(&otelHandler{inner: base})
}

// otelHandler wraps any slog.Handler, injecting trace_id and span_id
// extracted from the OpenTelemetry span in the context.
type otelHandler struct {
	inner slog.Handler
}

func (h *otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()))
	}
	if sc.HasSpanID() {
		r.AddAttrs(slog.String("span_id", sc.SpanID().String()))
	}
	return h.inner.Handle(ctx, r)
}

func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{inner: h.inner.WithGroup(name)}
}
