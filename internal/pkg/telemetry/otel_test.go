package telemetry

import (
	"bytes"
	"context"
	"testing"
)

func TestInit_StdoutFallback(t *testing.T) {
	var buf bytes.Buffer
	result, err := Init(context.Background(), Config{
		ServiceName:  "test",
		TraceWriter:  &buf,
		MetricWriter: &buf,
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}
	if result.PrometheusHandler != nil {
		t.Fatal("expected nil PrometheusHandler when Prometheus is disabled")
	}
	if err := result.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestInit_PrometheusEnabled(t *testing.T) {
	result, err := Init(context.Background(), Config{
		ServiceName:       "test-prom",
		PrometheusEnabled: true,
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if result.PrometheusHandler == nil {
		t.Fatal("expected non-nil PrometheusHandler when Prometheus is enabled")
	}
	if err := result.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestInit_DefaultServiceName(t *testing.T) {
	result, err := Init(context.Background(), Config{})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := result.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}
