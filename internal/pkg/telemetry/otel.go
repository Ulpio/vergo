// Package telemetry provides OpenTelemetry SDK setup for tracing and metrics.
package telemetry

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config holds optional settings for telemetry initialization.
type Config struct {
	ServiceName    string
	ServiceVersion string
	TraceWriter    io.Writer // nil = discard (e.g. when OTLP is used)
	MetricWriter   io.Writer // nil = discard
}

// Init initializes the global TracerProvider and MeterProvider with stdout
// exporters (or custom writers). Returns a shutdown function that must be
// called before process exit to flush and shut down both providers.
func Init(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "vergo"
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// TracerProvider
	var traceExp sdktrace.SpanExporter
	if cfg.TraceWriter != nil {
		traceExp, err = stdouttrace.New(stdouttrace.WithWriter(cfg.TraceWriter))
	} else {
		traceExp, err = stdouttrace.New(stdouttrace.WithWriter(io.Discard))
	}
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	// MeterProvider
	var metricExp metric.Exporter
	if cfg.MetricWriter != nil {
		metricExp, err = stdoutmetric.New(stdoutmetric.WithWriter(cfg.MetricWriter))
	} else {
		metricExp, err = stdoutmetric.New(stdoutmetric.WithWriter(io.Discard))
	}
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExp,
			metric.WithInterval(10*time.Second),
		)),
	)
	otel.SetMeterProvider(mp)

	shutdown = func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var firstErr error
		if e := tp.Shutdown(shutdownCtx); e != nil {
			firstErr = e
		}
		if e := mp.Shutdown(shutdownCtx); e != nil && firstErr == nil {
			firstErr = e
		}
		return firstErr
	}
	return shutdown, nil
}
