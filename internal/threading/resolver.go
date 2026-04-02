package threading

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

const (
	MatchStrategyInReplyTo  = "in_reply_to"
	MatchStrategyReferences = "references"
	MatchStrategyNewThread  = "new_thread"
)

type ThreadLookupRepository interface {
	FindByMessageID(ctx context.Context, messageID string) (domain.Thread, bool, error)
}

type Resolver struct {
	lookup ThreadLookupRepository
}

func NewResolver(lookup ThreadLookupRepository) Resolver {
	return Resolver{lookup: lookup}
}

func (r Resolver) Resolve(ctx context.Context, input core.ThreadResolutionInput) (core.ThreadResolutionResult, error) {
	if input.ParsedMessage.InReplyTo != "" {
		thread, found, err := r.lookup.FindByMessageID(ctx, input.ParsedMessage.InReplyTo)
		if err != nil {
			return core.ThreadResolutionResult{}, err
		}
		if found {
			return core.ThreadResolutionResult{
				Thread:    thread,
				MatchedBy: MatchStrategyInReplyTo,
				Created:   false,
			}, nil
		}
	}

	for _, reference := range input.ParsedMessage.References {
		thread, found, err := r.lookup.FindByMessageID(ctx, reference)
		if err != nil {
			return core.ThreadResolutionResult{}, err
		}
		if found {
			return core.ThreadResolutionResult{
				Thread:    thread,
				MatchedBy: MatchStrategyReferences,
				Created:   false,
			}, nil
		}
	}

	return core.ThreadResolutionResult{
		Thread: domain.Thread{
			ID:                newThreadID(),
			OrganizationID:    input.OrganizationID,
			AgentID:           input.AgentID,
			InboxID:           input.InboxID,
			ContactID:         input.ContactID,
			SubjectNormalized: input.ParsedMessage.SubjectNormalized,
			State:             domain.ThreadStateActive,
			LastActivityAt:    input.ReceivedAt,
			CreatedAt:         input.ReceivedAt,
			UpdatedAt:         input.ReceivedAt,
		},
		MatchedBy: MatchStrategyNewThread,
		Created:   true,
	}, nil
}

func newThreadID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "thread-generated"
	}
	return "thread-" + hex.EncodeToString(buf[:])
}
