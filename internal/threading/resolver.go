package threading

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
)

const (
	MatchStrategyInReplyTo     = "in_reply_to"
	MatchStrategyReferences    = "references"
	MatchStrategySubjectRecent = "subject_recent"
	MatchStrategyNewThread     = "new_thread"
)

type ThreadLookupRepository interface {
	FindByMessageID(ctx context.Context, messageID string) (domain.Thread, bool, error)
	FindMostRecentBySubject(ctx context.Context, organizationID, inboxID, contactID, subjectNormalized string) (domain.Thread, bool, error)
}

type Config struct {
	DormantThreshold time.Duration
}

type Resolver struct {
	lookup ThreadLookupRepository
	config Config
}

func NewResolver(lookup ThreadLookupRepository) Resolver {
	return NewResolverWithConfig(lookup, Config{
		DormantThreshold: 30 * 24 * time.Hour,
	})
}

func NewResolverWithConfig(lookup ThreadLookupRepository, config Config) Resolver {
	if config.DormantThreshold <= 0 {
		config.DormantThreshold = 30 * 24 * time.Hour
	}

	return Resolver{
		lookup: lookup,
		config: config,
	}
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

	if input.ParsedMessage.SubjectNormalized != "" {
		thread, found, err := r.lookup.FindMostRecentBySubject(
			ctx,
			input.OrganizationID,
			input.InboxID,
			input.ContactID,
			input.ParsedMessage.SubjectNormalized,
		)
		if err != nil {
			return core.ThreadResolutionResult{}, err
		}
		if found && !r.isDormant(thread, input.ReceivedAt) {
			return core.ThreadResolutionResult{
				Thread:    thread,
				MatchedBy: MatchStrategySubjectRecent,
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

func (r Resolver) isDormant(thread domain.Thread, receivedAt time.Time) bool {
	if thread.State == domain.ThreadStateDormant {
		return true
	}
	if thread.LastActivityAt.IsZero() {
		return false
	}
	return receivedAt.Sub(thread.LastActivityAt) > r.config.DormantThreshold
}

func newThreadID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "thread-generated"
	}
	return "thread-" + hex.EncodeToString(buf[:])
}
