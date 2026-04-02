package contacts

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type Repository interface {
	FindByEmail(ctx context.Context, organizationID, emailAddress string) (domain.Contact, bool, error)
}

type Resolver struct {
	repository Repository
}

func NewResolver(repository Repository) Resolver {
	return Resolver{repository: repository}
}

func (r Resolver) Resolve(ctx context.Context, input core.ContactResolutionInput) (core.ContactResolutionResult, error) {
	existing, found, err := r.repository.FindByEmail(ctx, input.OrganizationID, input.ParsedMessage.From.Email)
	if err != nil {
		return core.ContactResolutionResult{}, err
	}

	if found {
		if input.ParsedMessage.From.DisplayName != "" {
			existing.DisplayName = input.ParsedMessage.From.DisplayName
		}
		existing.LastSeenAt = input.ReceivedAt
		existing.UpdatedAt = input.ReceivedAt

		return core.ContactResolutionResult{
			Contact: existing,
			Created: false,
		}, nil
	}

	return core.ContactResolutionResult{
		Contact: domain.Contact{
			ID:             newContactID(),
			OrganizationID: input.OrganizationID,
			EmailAddress:   input.ParsedMessage.From.Email,
			DisplayName:    input.ParsedMessage.From.DisplayName,
			LastSeenAt:     input.ReceivedAt,
			CreatedAt:      input.ReceivedAt,
			UpdatedAt:      input.ReceivedAt,
		},
		Created: true,
	}, nil
}

func newContactID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "contact-generated"
	}
	return "contact-" + hex.EncodeToString(buf[:])
}
