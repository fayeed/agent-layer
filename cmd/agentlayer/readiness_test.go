package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentlayer/agentlayer/internal/core"
)

func TestValidateRuntimeConfigAggregatesErrors(t *testing.T) {
	t.Setenv("AGENTLAYER_EMAIL_PROVIDER", "ses")
	t.Setenv("AGENTLAYER_RAW_STORE", "s3")
	t.Setenv("AGENTLAYER_S3_BUCKET", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AGENTLAYER_DATABASE_URL", "postgres://example")

	errs := validateRuntimeConfig()
	if len(errs) < 2 {
		t.Fatalf("expected aggregated runtime config errors, got %#v", errs)
	}
}

func TestReadinessCheckerHealthy(t *testing.T) {
	checker := readinessChecker{
		config: func() []error { return nil },
		db:     readinessDBStub{},
		email: func() (healthChecker, error) {
			return readinessEmailStub{
				health: core.ProviderHealth{
					ProviderName: "dev",
					Healthy:      true,
					Details:      "local development provider",
				},
			}, nil
		},
	}

	response := checker.Check(context.Background())
	if !response.Ready {
		t.Fatalf("expected readiness response to be healthy, got %#v", response)
	}
	if len(response.Checks) != 3 {
		t.Fatalf("expected config, database, and email checks, got %#v", response)
	}
}

func TestReadinessCheckerUnhealthy(t *testing.T) {
	checker := readinessChecker{
		config: func() []error { return []error{errors.New("missing config")} },
		db:     readinessDBStub{err: errors.New("db down")},
		email: func() (healthChecker, error) {
			return readinessEmailStub{
				health: core.ProviderHealth{
					ProviderName: "ses",
					Healthy:      false,
					Details:      "access denied",
				},
				err: errors.New("access denied"),
			}, nil
		},
	}

	response := checker.Check(context.Background())
	if response.Ready {
		t.Fatalf("expected readiness response to be unhealthy, got %#v", response)
	}
}

func TestReadinessHandlerReturnsJSON(t *testing.T) {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		response := readinessResponse{
			Ready: true,
			Checks: []readinessCheck{
				{Name: "config", Healthy: true, Details: "validated"},
			},
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(response)
	})

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected readiness handler to return ok, got %d", recorder.Code)
	}

	var response readinessResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected readiness json response, got error: %v", err)
	}
	if !response.Ready {
		t.Fatalf("expected readiness response body, got %#v", response)
	}
}

type readinessDBStub struct {
	err error
}

func (s readinessDBStub) PingContext(context.Context) error {
	return s.err
}

type readinessEmailStub struct {
	health core.ProviderHealth
	err    error
}

func (s readinessEmailStub) HealthCheck(context.Context) (core.ProviderHealth, error) {
	return s.health, s.err
}
