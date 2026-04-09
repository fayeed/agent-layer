package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
)

type healthChecker interface {
	HealthCheck(ctx context.Context) (core.ProviderHealth, error)
}

type readinessCheck struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Details string `json:"details,omitempty"`
}

type readinessResponse struct {
	Ready  bool             `json:"ready"`
	Checks []readinessCheck `json:"checks"`
}

type readinessChecker struct {
	config func() []error
	db     interface {
		PingContext(ctx context.Context) error
	}
	email func() (healthChecker, error)
}

func newReadinessChecker() readinessChecker {
	checker := readinessChecker{
		config: validateRuntimeConfig,
		email: func() (healthChecker, error) {
			return newEmailProvider(), nil
		},
	}
	if store, ok := runtimeStore.(interface {
		PingContext(ctx context.Context) error
	}); ok {
		checker.db = store
	}
	return checker
}

func validateRuntimeConfig() []error {
	var errs []error
	if err := validateEmailProviderConfig(); err != nil {
		errs = append(errs, err)
	}
	if databaseURL() != "" {
		if err := validateRawStoreConfig(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c readinessChecker) Check(ctx context.Context) readinessResponse {
	checks := make([]readinessCheck, 0, 3)
	ready := true

	if errs := c.config(); len(errs) > 0 {
		ready = false
		for _, err := range errs {
			checks = append(checks, readinessCheck{
				Name:    "config",
				Healthy: false,
				Details: err.Error(),
			})
		}
	} else {
		checks = append(checks, readinessCheck{
			Name:    "config",
			Healthy: true,
			Details: "validated",
		})
	}

	if c.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := c.db.PingContext(dbCtx); err != nil {
			ready = false
			checks = append(checks, readinessCheck{
				Name:    "database",
				Healthy: false,
				Details: err.Error(),
			})
		} else {
			checks = append(checks, readinessCheck{
				Name:    "database",
				Healthy: true,
				Details: "reachable",
			})
		}
	}

	if c.email != nil {
		provider, err := c.email()
		if err != nil {
			ready = false
			checks = append(checks, readinessCheck{
				Name:    "email_provider",
				Healthy: false,
				Details: err.Error(),
			})
		} else {
			emailCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			health, err := provider.HealthCheck(emailCtx)
			if err != nil || !health.Healthy {
				ready = false
				details := health.Details
				if details == "" && err != nil {
					details = err.Error()
				}
				checks = append(checks, readinessCheck{
					Name:    "email_provider",
					Healthy: false,
					Details: details,
				})
			} else {
				checks = append(checks, readinessCheck{
					Name:    "email_provider",
					Healthy: true,
					Details: health.Details,
				})
			}
		}
	}

	return readinessResponse{
		Ready:  ready,
		Checks: checks,
	}
}

func newReadinessHandler() http.Handler {
	checker := newReadinessChecker()
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		response := checker.Check(request.Context())
		writer.Header().Set("Content-Type", "application/json")
		if response.Ready {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(writer).Encode(response)
	})
}
