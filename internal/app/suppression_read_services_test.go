package app

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestSuppressionReadServiceGetsSuppression(t *testing.T) {
	service := NewSuppressionReadService(suppressionGetterStub{
		record: domain.SuppressedAddress{ID: "suppression-123"},
	})

	record, err := service.GetSuppression(context.Background(), "suppression-123")
	if err != nil {
		t.Fatalf("expected suppression read to succeed, got error: %v", err)
	}
	if record.ID != "suppression-123" {
		t.Fatalf("expected suppression record, got %#v", record)
	}
}

func TestSuppressionListServiceListsSuppressions(t *testing.T) {
	service := NewSuppressionListService(suppressionListerStub{
		records: []domain.SuppressedAddress{{ID: "suppression-123"}},
	}, 20)

	records, err := service.ListSuppressions(context.Background(), 0)
	if err != nil {
		t.Fatalf("expected suppression list to succeed, got error: %v", err)
	}
	if len(records) != 1 || records[0].ID != "suppression-123" {
		t.Fatalf("expected suppression records, got %#v", records)
	}
}

type suppressionGetterStub struct {
	record domain.SuppressedAddress
	err    error
}

func (s suppressionGetterStub) GetSuppressionByID(context.Context, string) (domain.SuppressedAddress, error) {
	return s.record, s.err
}

type suppressionListerStub struct {
	records []domain.SuppressedAddress
	err     error
}

func (s suppressionListerStub) ListSuppressions(context.Context, int) ([]domain.SuppressedAddress, error) {
	return s.records, s.err
}
