# Observability

The services are instrumented with [OpenTelemetry](https://opentelemetry.io/) — one SDK emitting traces and metrics over OTLP gRPC. Logs are structured JSON (slog) with `trace_id`/`span_id` injected so logs and spans can be correlated.

## Stack

| Concern | Tool | Notes |
|---|---|---|
| Instrumentation | OpenTelemetry Go SDK + `otelgrpc` stats handlers | Auto spans + RED metrics for every RPC, server and client side |
| Traces | Jaeger (all-in-one) | OTLP gRPC receiver on :4317 |
| Metrics | Prometheus | OTLP receiver enabled, resource attributes promoted |
| Logs | Loki + Grafana Alloy | Alloy ships every container's stdout |
| Dashboards | Grafana | Provisioned datasources + gRPC RED dashboard |

## Run it

```sh
make obs
```

or individually: `make docker-up` (app) and `make observability-up` (stack).

## UIs & ports

| Service | URL |
|---|---|
| Grafana (dashboards) | http://localhost:3000 (admin/admin) |
| Jaeger (spans) | http://localhost:16686 |
| Prometheus | http://localhost:9090 |
| Loki | http://localhost:3100 |
| Alloy | http://localhost:12345 |

`make observability-down` / `make observability-logs` for the stack lifecycle.

## What to look at

Opening http://localhost:3000 lands directly on the **gRPC RED Metrics** dashboard (set as the default home dashboard via `GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH`).

Generate data first: `make traffic` (60s of requests so the panels light up).

- **Traces**: open Jaeger → service `bookie-client` → hit `http://localhost:8080/books/1234` → a full client→server trace appears (`Bookie/GetByID` spans under both `bookie-client` and `bookie-grpc`).
- **Metrics**: the RED dashboard — request rate, error rate (by gRPC status code), p50/p95/p99. Prometheus translates OTLP into standard `rpc_*` PromQL, e.g. `rpc_server_call_duration_seconds_count{service_name="bookie-grpc", rpc_response_status_code="OK"}`.
- **Logs**: Grafana → Explore → Loki. App records are JSON and promoted to the `service` label; `trace_id`/`span_id` appear in log lines when the context carries a span.

## Configuration

Standard `OTEL_*` env vars are honored by the app (see `src/internal/utils/telemetry.go`):

- `OTEL_EXPORTER_OTLP_ENDPOINT` — OTLP gRPC receiver for traces (default `localhost:4317`; set to `jaeger:4317` inside compose).
- `OTEL_METRICS_OTLP_ENDPOINT` — OTLP HTTP receiver for metrics (default `localhost`; set to `prometheus:9090` inside compose). Exported to `/api/v1/otlp/v1/metrics`.

## Files

- `src/internal/utils/telemetry.go` — provider setup + `ShutdownTelemetry` (wired into graceful shutdown)
- `src/internal/utils/trace_logger.go` — slog handler enriching logs with trace context
- `observability/` — Prometheus, Grafana, Loki, Alloy configs
- `docker-compose.yml` — `jaeger`, `prometheus`, `loki`, `alloy`, `grafana` services
