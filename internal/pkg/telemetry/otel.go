// Package telemetry provides OpenTelemetry SDK setup for tracing and metrics.
package telemetry

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds optional settings for telemetry initialization.
type Config struct {
	ServiceName    string
	ServiceVersion string
	TraceWriter    io.Writer // nil = discard (used when no OTLP endpoint)
	MetricWriter   io.Writer // nil = discard (used when Prometheus is disabled)

	// OTLP exporter settings (empty = use stdout fallback)
	OTLPEndpoint string // gRPC endpoint, e.g. "localhost:4317"
	OTLPInsecure bool   // use insecure connection (no TLS)

	// Prometheus metrics (when enabled, exposes /metrics for scraping)
	PrometheusEnabled bool
}

// Result holds references needed after Init (e.g. Prometheus handler).
type Result struct {
	Shutdown          func(context.Context) error
	PrometheusHandler *prometheus.Exporter // nil when Prometheus is disabled
}

// Init initializes the global TracerProvider and MeterProvider.
// When OTLPEndpoint is set, traces are exported via OTLP gRPC (e.g. to Jaeger).
// When PrometheusEnabled is true, metrics are exposed via a Prometheus scrape endpoint.
// Falls back to stdout exporters when neither is configured.
func Init(ctx context.Context, cfg Config) (*Result, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "vergo"
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
		resource.WithProcessRuntimeDescription(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	// --- TracerProvider ---
	traceExp, err := newTraceExporter(ctx, cfg)
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

	// --- MeterProvider ---
	mp, promExporter, err := newMeterProvider(res, cfg)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}
	otel.SetMeterProvider(mp)

	shutdown := func(ctx context.Context) error {
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

	return &Result{
		Shutdown:          shutdown,
		PrometheusHandler: promExporter,
	}, nil
}

// newTraceExporter returns an OTLP gRPC exporter when endpoint is set,
// otherwise falls back to stdout.
func newTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint != "" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}
		if cfg.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(ctx, opts...)
	}

	// Fallback: stdout (discard when no writer)
	w := cfg.TraceWriter
	if w == nil {
		w = io.Discard
	}
	return stdouttrace.New(stdouttrace.WithWriter(w))
}

// newMeterProvider returns a MeterProvider with Prometheus exporter when enabled,
// otherwise falls back to stdout periodic reader.
func newMeterProvider(res *resource.Resource, cfg Config) (*metric.MeterProvider, *prometheus.Exporter, error) {
	if cfg.PrometheusEnabled {
		promExporter, err := prometheus.New()
		if err != nil {
			return nil, nil, err
		}
		mp := metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(promExporter),
		)
		return mp, promExporter, nil
	}

	// Fallback: stdout periodic reader
	w := cfg.MetricWriter
	if w == nil {
		w = io.Discard
	}
	metricExp, err := stdoutmetric.New(stdoutmetric.WithWriter(w))
	if err != nil {
		return nil, nil, err
	}
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExp,
			metric.WithInterval(10*time.Second),
		)),
	)
	return mp, nil, nil
}
