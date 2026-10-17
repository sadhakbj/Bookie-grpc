---
status: accepted
---

# App logs travel via OTLP push, not docker-socket scraping

Observability was previously coupled to *where the app runs*: logs reached Loki only by
scraping docker container stdout through `/var/run/docker.sock`. That breaks in k8s/GKE
(pod has no docker socket; mounting one is an antipattern) and silently misses native
`make serve` / `make client` processes. We decided first-party apps export logs over
OTLP push (`otel/sdk/log` + `otelslog` bridge, to a dedicated `OTLP_LOGS_OTLP_ENDPOINT`
validated at bootstrap), and the observability stack becomes a pure receiver: Alloy
serves the OTLP log receiver and routes into Loki; docker-socket scraping plus its host
mount are retired.

## Considered alternatives

- **Keep scraping + shims per platform**: a file/stream shim in `make serve`, a
  hostPath mount in k8s. Rejected — one platform-specific hack per runtime, and the k8s
  variant stays broken by design.
- **sidecar ssh-agent style docker socket access in k8s**: rejected — the antipattern
  this decision removes.

## Consequences

- The telemetry contract is uniform across `make serve`, docker-compose, and k8s/GKE:
  traces, metrics, and logs all go out over one OTLP endpoint, boot-time validated.
- Unreachable observability degrades to dropped batches — identical to traces today; no
  new app↔obs coupling is introduced (still none at init).
- Through OTLP, log records inherit the active trace context — Loki records now carry
  `trace_id` for native runs too, not just containers.
- Loki no longer sees generic third-party container stdout with the docker source
  removed; re-adding it later is additive and doesn't touch the app.
- Alloy no longer needs the docker.sock mount — one less privilege.
