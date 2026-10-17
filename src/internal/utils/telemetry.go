// Package utils provides telemetry setup (traces + metrics) for the services.
package utils

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

var (
	shutdownOnce sync.Once
	shutdownFns  []func(context.Context) error
)

// TelemetryConfig contains the exporter endpoints used by telemetry providers.
type TelemetryConfig struct {
	OTLPEndpoint        string
	OTLPMetricsEndpoint string
}

// InitializeTelemetry wires OpenTelemetry tracer and meter providers.
// Traces go over OTLP gRPC (Jaeger), metrics over OTLP HTTP to the
// Prometheus OTLP receiver. It is safe to call once per process.
func InitializeTelemetry(ctx context.Context, logger *slog.Logger, serviceName string, config TelemetryConfig) error {
	res, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewSchemaless(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return err
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
	)
	if err != nil {
		return err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithURLPath("/api/v1/otlp/v1/metrics"),
	}
	metricOpts = append(metricOpts, otlpmetrichttp.WithEndpoint(config.OTLPMetricsEndpoint))
	metricExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(15*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)

	shutdownFns = append(shutdownFns,
		func(ctx context.Context) error { return tracerProvider.Shutdown(ctx) },
		func(ctx context.Context) error { return meterProvider.Shutdown(ctx) },
	)

	logger.Info("Telemetry initialized",
		"service", serviceName,
		"otlp_endpoint", config.OTLPEndpoint,
		"metrics_endpoint", config.OTLPMetricsEndpoint,
	)
	return nil
}

// ShutdownTelemetry flushes and stops all telemetry providers. It is safe to
// call multiple times; only the first call performs the flush.
func ShutdownTelemetry(ctx context.Context, logger *slog.Logger) {
	shutdownOnce.Do(func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		for _, fn := range shutdownFns {
			if err := fn(shutdownCtx); err != nil {
				logger.Error("Failed to shut down telemetry provider: %v", "error", err)
			}
		}
		logger.Info("Telemetry shut down")
	})
}
