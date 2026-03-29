package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestOtelHandler_WithSpanContext(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(&otelHandler{inner: base})

	traceID, _ := trace.TraceIDFromHex("0af7651916cd43dd8448eb211c80319c")
	spanID, _ := trace.SpanIDFromHex("b7ad6b7169203331")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.InfoContext(ctx, "test message")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got := entry["trace_id"]; got != "0af7651916cd43dd8448eb211c80319c" {
		t.Errorf("trace_id = %v, want 0af7651916cd43dd8448eb211c80319c", got)
	}
	if got := entry["span_id"]; got != "b7ad6b7169203331" {
		t.Errorf("span_id = %v, want b7ad6b7169203331", got)
	}
	if got := entry["msg"]; got != "test message" {
		t.Errorf("msg = %v, want test message", got)
	}
}

func TestOtelHandler_WithoutSpan(t *testing.T) {
	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(&otelHandler{inner: base})

	logger.InfoContext(context.Background(), "no span")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := entry["trace_id"]; ok {
		t.Error("trace_id should not be present without span context")
	}
	if _, ok := entry["span_id"]; ok {
		t.Error("span_id should not be present without span context")
	}
}

func TestNew_DevMode(t *testing.T) {
	logger := New("dev")
	if logger == nil {
		t.Fatal("New(dev) returned nil")
	}
	// Verify it can log without panic
	logger.Info("dev mode test")
}

func TestNew_ProdMode(t *testing.T) {
	logger := New("production")
	if logger == nil {
		t.Fatal("New(production) returned nil")
	}
	logger.Info("prod mode test")
}
