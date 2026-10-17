// Package utils provides telemetry setup (traces + metrics + logs) for the services.
package utils

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
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
	OTLPLogsEndpoint    string
	OTLPMetricsEndpoint string
}

// InitializeTelemetry wires OpenTelemetry tracer, meter and logger providers and
// returns the logger that mirrors records to OTLP when OTLPLogsEndpoint is set.
// Exporters are fire-and-forget: an unreachable endpoint drops batches without
// affecting application behavior. It is safe to call once per process.
func InitializeTelemetry(ctx context.Context, logger *slog.Logger, serviceName string, config TelemetryConfig) (*slog.Logger, error) {
	res, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewSchemaless(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return logger, err
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
	)
	if err != nil {
		return logger, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	shutdownFns = append(shutdownFns, tracerProvider.Shutdown)

	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(config.OTLPMetricsEndpoint),
		otlpmetrichttp.WithURLPath("/api/v1/otlp/v1/metrics"),
	)
	if err != nil {
		return logger, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(15*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)
	shutdownFns = append(shutdownFns, meterProvider.Shutdown)

	readyLogger := logger
	if config.OTLPLogsEndpoint != "" {
		logExporter, err := otlploggrpc.New(ctx,
			otlploggrpc.WithInsecure(),
			otlploggrpc.WithEndpoint(config.OTLPLogsEndpoint),
		)
		if err != nil {
			return logger, err
		}

		loggerProvider := sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		)
		shutdownFns = append(shutdownFns, loggerProvider.Shutdown)

		readyLogger = slog.New(newSlogFanout(
			logger.Handler(),
			otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider)),
		))
	}

	readyLogger.Info("Telemetry initialized",
		"otlp_endpoint", config.OTLPEndpoint,
		"logs_endpoint", config.OTLPLogsEndpoint,
		"metrics_endpoint", config.OTLPMetricsEndpoint,
	)
	return readyLogger, nil
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
