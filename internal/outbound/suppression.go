package outbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type SuppressionRepository interface {
	Save(ctx context.Context, record domain.SuppressedAddress) (domain.SuppressedAddress, error)
}

type SuppressionInput struct {
	Message    domain.Message
	Contact    domain.Contact
	Status     string
	OccurredAt time.Time
}

type SuppressionService struct {
	repository SuppressionRepository
}

func NewSuppressionService(repository SuppressionRepository) SuppressionService {
	return SuppressionService{repository: repository}
}

func (s SuppressionService) Apply(ctx context.Context, input SuppressionInput) (domain.SuppressedAddress, bool, error) {
	if input.Status != DeliveryStateHardBounce && input.Status != DeliveryStateComplaint {
		return domain.SuppressedAddress{}, false, nil
	}

	record := domain.SuppressedAddress{
		ID:             newSuppressionID(),
		OrganizationID: input.Message.OrganizationID,
		EmailAddress:   input.Contact.EmailAddress,
		Reason:         input.Status,
		Source:         "provider_callback",
		CreatedAt:      input.OccurredAt,
		UpdatedAt:      input.OccurredAt,
	}

	saved, err := s.repository.Save(ctx, record)
	if err != nil {
		return domain.SuppressedAddress{}, false, err
	}

	return saved, true, nil
}

func newSuppressionID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "suppression-generated"
	}
	return "suppression-" + hex.EncodeToString(buf[:])
}
