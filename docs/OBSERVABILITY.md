<!-- last-reviewed: 2026-02-15 -->
# Observability

Local observability stack for querying logs, metrics, and traces produced by the application.

## Architecture

```
Application
    │
    │  OTLP HTTP (:4318)
    │  (traces, metrics, logs)
    ▼
  Vector
    │
    ├──► Victoria Logs   (:9428)  — LogsQL
    ├──► Victoria Metrics (:8428) — PromQL
    └──► Grafana Tempo    (:3200) — TraceQL

  Grafana (:3000) ─── queries all three backends
```

The application sends all three telemetry signals (traces, metrics, logs) via OTLP HTTP to Vector. Vector fans out to three storage backends. Grafana provides a unified UI for exploring all signals with pre-configured datasources.

## Quick Start

```bash
make observability-up     # start the stack
make observability-down   # tear it down
make observability-logs   # tail Vector logs
```

The stack is fully ephemeral — no persistent volumes. Tear down and rebuild cleanly.

## Service Endpoints

| Service | Port | Purpose | Query Language |
|---------|------|---------|----------------|
| Vector | `localhost:4318` | OTLP HTTP receiver (traces, metrics, logs) | — |
| Vector API | `localhost:8686` | Vector health and internal metrics | — |
| Victoria Logs | `localhost:9428` | Log storage and querying | LogsQL |
| Victoria Metrics | `localhost:8428` | Metric storage and querying | PromQL |
| Grafana Tempo | `localhost:3200` | Trace storage and querying | TraceQL |
| Grafana | `localhost:3000` | Visualization UI for all signals | — |

## Querying

### Logs (LogsQL)

```bash
# All logs from a service (replace <service> with your app's OTel service_name)
curl -s 'http://localhost:9428/select/logsql/query' \
  --data-urlencode 'query=service:<service>'

# Logs containing "error"
curl -s 'http://localhost:9428/select/logsql/query' \
  --data-urlencode 'query=service:<service> AND error'
```

### Metrics (PromQL)

```bash
# Query a metric
curl -s 'http://localhost:8428/api/v1/query' \
  --data-urlencode 'query=http_server_request_duration_seconds_count'

# Range query over last 5 minutes
curl -s 'http://localhost:8428/api/v1/query_range' \
  --data-urlencode 'query=rate(http_server_request_duration_seconds_count[5m])' \
  --data-urlencode 'start=now-5m' \
  --data-urlencode 'end=now' \
  --data-urlencode 'step=15s'
```

### Traces (TraceQL)

```bash
# Find traces by service name (replace <service> with your app's OTel service_name)
curl -s 'http://localhost:3200/api/search' \
  --data-urlencode 'q={resource.service.name="<service>"}'

# Find traces with errors
curl -s 'http://localhost:3200/api/search' \
  --data-urlencode 'q={resource.service.name="<service>" && status=error}'

# Get a specific trace by ID
curl -s 'http://localhost:3200/api/traces/<trace-id>'
```

## Configuration

- **Vector config:** `development/observability/vector.yaml`
- **Tempo config:** `development/observability/tempo.yaml`
- **Grafana provisioning:** `development/observability/grafana/provisioning/`
- **Docker Compose:** `development/docker-compose.observability.yml`
- **Application OTel config:** `apps/<app>/resources/config/development.yaml` (otel section)

## Application Wiring

OTel is configured via the `otelfx` module. When `enabled: true`, the `endpoint` field is **required** — validation fails at startup if it's empty. In development, `development.yaml` points to the local Vector endpoint:

```yaml
otel:
  enabled: true
  endpoint: localhost:4318
  service_name: <app>
  sample_rate: 1.0
  insecure: true
```

The `endpoint` field is `host:port` (not a full URL). The SDK prepends the protocol based on the `insecure` flag (`http://` when true, `https://` when false).

### Log Bridge

By default, slog writes only to stdout. To also export structured logs via OTLP (so they appear in VictoriaLogs alongside traces and metrics), enable the log bridge in the logging config:

```yaml
logging:
  level: info
  format: text
  otel_bridge: true
```

When enabled, the `loggerfx` module creates a fan-out handler that writes every log record to both stdout and the OTLP log exporter. Logs exported via OTLP include `trace_id` and `span_id` from the active span context, enabling trace-to-log correlation in Grafana.

### Automatic Instrumentation

Instrumentation is automatic for:
- **HTTP requests** — `otelhttp` middleware creates spans per request, bridges request ID and correlation ID as span attributes
- **Database queries** — `otelpgx` tracer creates spans for every pgx query

### Grafana

Grafana starts with anonymous admin access (no login) and pre-configured datasources for all three backends. Open `http://localhost:3000` and use **Explore** to query any signal.

Configuration is provisioned from `development/observability/grafana/provisioning/`.
