# Observability

Everything is provisioned from code — clones, `make obs`, and it works. No manual UI setup, ever.

The services are instrumented with [OpenTelemetry](https://opentelemetry.io/) — one SDK emitting traces and metrics over OTLP. Logs are structured JSON (slog) with `trace_id`/`span_id` injected so logs and spans can be correlated.

## Stack

| Concern | Tool | Notes |
|---|---|---|
| Instrumentation | OpenTelemetry Go SDK + `otelgrpc` stats handlers | Auto spans + RED metrics for every RPC, server and client side — no per-handler code |
| Traces | Jaeger (all-in-one) | OTLP gRPC receiver on :4317 |
| Metrics | Prometheus | OTLP receiver enabled, translated to standard `rpc_*` PromQL |
| Logs | Loki + Grafana Alloy | Alloy ships every container's stdout; entries older than 24h are dropped (Docker replays stale history on startup) |
| Dashboards | Grafana | Datasources + gRPC RED dashboard + live logs panel provisioned at startup |

## Installation guide (fresh clone → running)

Prerequisites: Docker (any recent version). That's it.

```sh
git clone https://github.com/sadhakbj/bookie-grpc.git
cd bookie-grpc
make obs        # builds app images, starts app + Jaeger + Prometheus + Loki + Alloy + Grafana
make traffic    # optional: 60s of demo traffic so dashboards have data
```

Open **http://localhost:3000** — you land directly on the gRPC RED dashboard: metrics on top, live logs at the bottom, both filtered by the same service selector. Traces: **http://localhost:16686**.

Stop: `make docker-down` (app) / `make observability-down` (stack). Logs: `make docker-logs` / `make observability-logs`.

Only pre-existing old containers in your Docker history are affected by the 24h log-drop filter; app logs are live.

## UIs & ports

| Service | URL |
|---|---|
| Grafana (metrics + logs, home = RED dashboard) | http://localhost:3000 (admin/admin) |
| Jaeger (traces) | http://localhost:16686 |
| Prometheus | http://localhost:9090 |
| Loki | http://localhost:3100 |
| Alloy | http://localhost:12345 |

## What to look at

- **Metrics (home dashboard)** — request rate, error rate (by gRPC status code), p50/p95/p99, client-side rate/errors. PromQL under the hood: `rpc_server_call_duration_seconds_count{service_name="bookie-grpc", rpc_response_status_code="OK"}`, `histogram_quantile(..., rpc_server_call_duration_seconds_bucket)`.
- **Logs (bottom panel of the same dashboard)** — live Loki stream, no query to type. App records are JSON with the `service` field promoted to a label; `trace_id`/`span_id` appear in lines when the context carries a span.
- **Traces (Jaeger)** — hit `http://localhost:8080/books/0199e26a-2951-7000-8000-000000000001`, then find the trace: `Bookie/GetByID` spans under both `bookie-client` and `bookie-grpc`, chained via W3C trace context propagation.
- If you ever need raw LogQL: `{source="docker"} | json service | service="bookie-grpc"`.

## Configuration

Standard `OTEL_*` env vars are honored by the app (see `src/internal/utils/telemetry.go`):

- `OTEL_EXPORTER_OTLP_ENDPOINT` — OTLP gRPC receiver for traces (default `localhost:4317`; set to `jaeger:4317` inside compose).
- `OTEL_METRICS_OTLP_ENDPOINT` — OTLP HTTP receiver for metrics (default `localhost:4318`; set to `prometheus:9090` inside compose, wiring exports to `/api/v1/otlp/v1/metrics`).

Grafana-specific env in compose: `GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH` pins the RED dashboard as home; admin credentials via `GF_SECURITY_ADMIN_*`; anonymous Viewer access enabled for zero-click dev use.

## Files

- `src/internal/utils/telemetry.go` — provider setup (traces → OTLP gRPC, metrics → OTLP HTTP) + `ShutdownTelemetry` wired into graceful shutdown
- `src/internal/utils/trace_logger.go` — slog handler enriching logs with trace context
- `observability/grafana/provisioning/` — datasource + dashboard provisioning definitions
- `observability/grafana/dashboards/grpc-red.json` — the RED dashboard + logs panel (compiled Grafana format)
- `observability/prometheus/prometheus.yml` — OTLP receiver config
- `observability/loki/loki-config.yaml`, `observability/alloy/config.alloy` — log pipeline
- `docker-compose.yml` — `jaeger`, `prometheus`, `loki`, `alloy`, `grafana` services
