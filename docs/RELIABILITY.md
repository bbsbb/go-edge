<!-- last-reviewed: 2026-02-15 content-hash: 5efb157f -->
# Reliability

Reliability contracts and operational behavior.

## Health Probes

### Liveness: `/healthz`

- Always returns HTTP 200 with `{"status": "pass"}`
- No external dependency checks
- Purpose: tells Kubernetes the process is alive
- If this fails, the process is dead and should be restarted

### Readiness: `/readyz`

- Checks PostgreSQL connectivity via `db.PingContext()`
- Returns HTTP 200 when the database is reachable
- Returns HTTP 503 when the database is unreachable
- Purpose: tells Kubernetes whether the pod should receive traffic
- A failing readiness probe removes the pod from the service load balancer but does not restart it

### Design Rationale

Liveness must never check external dependencies. A database outage should not cause Kubernetes to restart application pods in a loop. Readiness is the correct probe for dependency health — it stops traffic routing without destroying the process.

## Graceful Shutdown

Shutdown is managed by Uber FX lifecycle hooks. The sequence:

1. **Signal received** — FX handles SIGINT and SIGTERM automatically
2. **HTTP listener closes** — no new connections accepted
3. **In-flight requests drain** — existing requests continue to completion
4. **Database connections close** — `db.Close()` releases the connection pool
5. **Process exits**

### Timeouts

| Timeout | Default | Source |
|---------|---------|--------|
| Read header | 5s | `config.HTTPServer.ReadHeaderTimeout` (max 120s) |
| Read | 10s | `config.HTTPServer.ReadTimeout` (max 120s) |
| Write | RequestTimeout + 5s | Derived from config |
| Idle | 120s | `config.HTTPServer.IdleTimeout` (max 600s) |
| Request middleware | RequestTimeout | `config.HTTPServer.RequestTimeout` |
| FX stop timeout | RequestTimeout + 10s | Set in the application's composition root (e.g., `cmd/server.go`) |

The FX stop timeout is the total window for shutdown. It must be larger than the write timeout to allow in-flight responses to complete. The chain is:

```
Request middleware timeout < Write timeout < FX stop timeout
     RequestTimeout       < RT + 5s       < RT + 10s
```

### Database Startup

On startup, the PostgreSQL connection is verified with a 5-second ping timeout. If the database is unreachable at boot, the application fails to start.

## Panic Recovery

The `middlewarefx.Recovery` middleware catches panics in HTTP handlers, logs the panic value and full stack trace via slog, and returns an RFC 9457 problem details response (HTTP 500). Enabled by default via `DefaultConfiguration()`. Can be disabled but strongly discouraged — a panic must never crash the server or leak internal details.

## Request Body Limits

The `middlewarefx.MaxBytes` middleware wraps every request body with `http.MaxBytesReader`. Default limit is 1 MB. Configurable via `max_request_body_bytes` in the middleware configuration or the `MAX_REQUEST_BODY_BYTES` environment variable. Enabled by default via `DefaultConfiguration()`. Can be disabled but strongly discouraged — an unbounded POST can exhaust server memory.

## SLA Thresholds

Performance constraints that agents and CI can validate against the observability stack. These thresholds apply when the application is running with the local observability stack (`make observability-up`).

| Metric | Threshold | Rationale |
|--------|-----------|-----------|
| Service startup | < 5s | FX boot + DB ping + migrations must complete quickly |
| HTTP request p99 latency | < 500ms | User-facing requests should feel responsive |
| Max span duration | < 2s | No single operation should block for this long |
| Health probe response | < 100ms | `/healthz` and `/readyz` must be near-instant |

Validate with `tools/scripts/validate-sla.sh` when the app and observability stack are running. Thresholds are aspirational until an application exists to measure against — define them now so they're enforced from day one.

## Error Handling

See [ARCHITECTURE.md — Domain Error Model](../ARCHITECTURE.md#domain-error-model) for the error code classification, HTTP status mapping, and error translation flow.
