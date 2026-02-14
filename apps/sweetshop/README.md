# Sweetshop

Demo application for the go-edge framework. Implements a multi-tenant product catalog and ordering system with PostgreSQL, Row-Level Security, and OpenTelemetry observability.

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- PostgreSQL client tools (optional, for direct DB access)

## Local Development

### 1. Start Infrastructure

Start PostgreSQL (creates both `app_sweetshop` and `test_app_sweetshop` databases):

```sh
# From repository root
docker compose -f development/docker-compose.default.yml up postgres -d
```

Start the observability stack (Vector, VictoriaLogs, VictoriaMetrics, Tempo, Grafana):

```sh
make observability-up
```

### 2. Run Migrations

```sh
# From apps/sweetshop/
make migrate
```

Other migration commands:

```sh
make migrate-reset    # Roll back all migrations and re-apply (development only)
make migrate-verify   # Verify database is on latest migration version
```

### 3. Seed Data

```sh
# From apps/sweetshop/
bash scripts/seed.sh
```

### 4. Run the Application

```sh
# From apps/sweetshop/
make run
```

The server starts on `http://localhost:8080`.

### 5. Send Requests

All requests require a tenant identifier. Use the `X-Organization-Slug` header:

```sh
# List products
curl -H "X-Organization-Slug: dev-shop" http://localhost:8080/products

# Create a product
curl -H "X-Organization-Slug: dev-shop" -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Chocolate Cake","category":"ice_cream","price_cents":999}'
```

A full smoke test script is available:

```sh
bash scripts/requests.sh
```

### 6. Observability

The app exports all three telemetry signals via OTLP to Vector, which routes them to dedicated backends:

| Signal  | Backend          | Port  |
|---------|------------------|-------|
| Logs    | VictoriaLogs     | 9428  |
| Metrics | VictoriaMetrics  | 8428  |
| Traces  | Tempo            | 3200  |

**Grafana** is available at `http://localhost:3000` with pre-configured datasources for all three backends. No login required.

- **Explore > VictoriaLogs** — query structured logs (LogsQL)
- **Explore > VictoriaMetrics** — query metrics (PromQL)
- **Explore > Tempo** — search and inspect traces

Logs include `trace_id` and `span_id` fields, enabling trace-to-log correlation in Grafana.

CLI queries are also available from the repository root:

```sh
make query-logs Q="sweetshop"
make query-metrics Q="http_server_request_duration_seconds_count"
make query-traces Q="sweetshop"
```

## Creating Migrations

```sh
# SQL migration
make migrate-create NAME=create_users TYPE=sql

# Go migration
make migrate-create NAME=seed_data TYPE=go
```

Migrations are stored in `internal/migrations/versions/`. Goose uses sequential numbering automatically.

## Running Tests

Tests use the `test_app_sweetshop` database on the same Postgres instance:

```sh
docker compose -f development/docker-compose.default.yml up postgres -d  # from repository root
make test  # from apps/sweetshop/
```

## Docker

### Build

```sh
# From repository root
make docker-build APP=sweetshop
```

### Run (fully containerized)

```sh
docker compose -f development/docker-compose.default.yml up
```

This starts PostgreSQL, runs migrations via an init container, then starts the application. The app sends telemetry to `host.docker.internal:4318` (the host's observability stack).

### Production

The Docker image accepts commands via the entrypoint:

```sh
docker run go-edge/sweetshop server           # default
docker run go-edge/sweetshop migrate up
docker run go-edge/sweetshop migrate verify
```

Production config uses `secret://` references resolved at runtime. See `resources/config/production.yaml`.

## Configuration

Configuration is loaded from `resources/config/` by default. Override with `CONFIG_PATH` env var.

| File | Purpose |
|---|---|
| `development.yaml` | Local development (text logging, OTLP log bridge, localhost DB, full sampling) |
| `testing.yaml` | Test runner (JSON logging, no OTel, test database) |
| `migrate/development.yaml` | Migration runner config (root credentials for schema changes) |
| `production.yaml` | Production (JSON logging, secret:// references, SSL, 10% sampling) |
| `migrate/production.yaml` | Production migration config (secret:// references, SSL) |

Environment variables override config values with the prefix `APP_SWEETSHOP_` (e.g., `APP_SWEETSHOP_PSQL_HOST=postgres`).
