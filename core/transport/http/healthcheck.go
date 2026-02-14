package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// See https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check
type HealthStatus string

const (
	HealthStatusPass HealthStatus = "pass"
	HealthStatusFail HealthStatus = "fail"
	HealthStatusWarn HealthStatus = "warn"
)

type ComponentType string

const (
	ComponentTypeDatastore ComponentType = "datastore"
	ComponentTypeSystem    ComponentType = "system"
)

type HealthCheck struct {
	ComponentType ComponentType `json:"component_type"`
	Status        HealthStatus  `json:"status"`
}

type HealthResponse struct {
	Status HealthStatus             `json:"status"`
	Checks map[string][]HealthCheck `json:"checks"`
}

func CheckPostgres(pool *pgxpool.Pool, timeout time.Duration) (HealthStatus, map[string][]HealthCheck) {
	check := HealthCheck{ComponentType: ComponentTypeDatastore}

	if timeout == 0 {
		timeout = 1 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if pool == nil {
		check.Status = HealthStatusFail
	} else if err := pool.Ping(ctx); err != nil {
		check.Status = HealthStatusFail
	} else {
		check.Status = HealthStatusPass
	}

	return check.Status, map[string][]HealthCheck{
		"postgres": {check},
	}
}

func LivenessHandler() http.HandlerFunc {
	resp := HealthResponse{Status: HealthStatusPass}
	body, _ := json.Marshal(resp)

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck,gosec
	}
}

func ReadinessHandler(pool *pgxpool.Pool, timeout time.Duration, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, checks := CheckPostgres(pool, timeout)
		resp := HealthResponse{Status: status, Checks: checks}

		w.Header().Set("Content-Type", "application/json")
		if status == HealthStatusFail {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode health response", "error", err)
		}
	}
}
