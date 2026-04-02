package webhooks

import (
	"context"
	"encoding/json"

	"github.com/agentlayer/agentlayer/internal/core"
	"github.com/agentlayer/agentlayer/internal/domain"
	"github.com/agentlayer/agentlayer/internal/inbound"
)

type BuildMessageReceivedInput struct {
	Organization   domain.Organization
	Agent          domain.Agent
	Inbox          domain.Inbox
	Delivery       domain.WebhookDelivery
	Handled        inbound.HandleResult
	ThreadMessages []domain.Message
	Memory         []domain.ContactMemoryEntry
}

type MessageReceivedPayload struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	OccurredAt     string                 `json:"occurred_at"`
	Organization   payloadOrganization    `json:"organization"`
	Agent          payloadAgent           `json:"agent"`
	Inbox          payloadInbox           `json:"inbox"`
	Message        payloadMessage         `json:"message"`
	Thread         payloadThread          `json:"thread"`
	Contact        payloadContact         `json:"contact"`
	ThreadMessages []payloadMessage       `json:"thread_messages"`
	Memory         []payloadContactMemory `json:"memory"`
}

type MessageReceivedBuilder struct{}

func NewMessageReceivedBuilder() MessageReceivedBuilder {
	return MessageReceivedBuilder{}
}

func (MessageReceivedBuilder) Build(_ context.Context, input BuildMessageReceivedInput) (core.WebhookDispatchRequest, error) {
	payload := MessageReceivedPayload{
		EventID:    input.Delivery.EventID,
		EventType:  input.Delivery.EventType,
		OccurredAt: input.Delivery.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Organization: payloadOrganization{
			ID:   input.Organization.ID,
			Name: input.Organization.Name,
		},
		Agent: payloadAgent{
			ID:   input.Agent.ID,
			Name: input.Agent.Name,
		},
		Inbox: payloadInbox{
			ID:           input.Inbox.ID,
			EmailAddress: input.Inbox.EmailAddress,
		},
		Message: mapMessage(input.Handled.Message),
		Thread: payloadThread{
			ID:                input.Handled.Thread.ID,
			SubjectNormalized: input.Handled.Thread.SubjectNormalized,
			State:             string(input.Handled.Thread.State),
		},
		Contact: payloadContact{
			ID:           input.Handled.Contact.ID,
			EmailAddress: input.Handled.Contact.EmailAddress,
			DisplayName:  input.Handled.Contact.DisplayName,
		},
		ThreadMessages: make([]payloadMessage, 0, len(input.ThreadMessages)),
		Memory:         make([]payloadContactMemory, 0, len(input.Memory)),
	}

	for _, message := range input.ThreadMessages {
		payload.ThreadMessages = append(payload.ThreadMessages, mapMessage(message))
	}

	for _, memory := range input.Memory {
		payload.Memory = append(payload.Memory, payloadContactMemory{
			ID:        memory.ID,
			ContactID: memory.ContactID,
			ThreadID:  memory.ThreadID,
			Note:      memory.Note,
			Tags:      memory.Tags,
			CreatedAt: memory.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return core.WebhookDispatchRequest{}, err
	}

	return core.WebhookDispatchRequest{
		Delivery: input.Delivery,
		Payload:  body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

type payloadOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type payloadAgent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type payloadInbox struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
}

type payloadMessage struct {
	ID              string `json:"id"`
	ThreadID        string `json:"thread_id"`
	Direction       string `json:"direction"`
	Subject         string `json:"subject"`
	TextBody        string `json:"text_body"`
	MessageIDHeader string `json:"message_id_header"`
	CreatedAt       string `json:"created_at"`
}

type payloadThread struct {
	ID                string `json:"id"`
	SubjectNormalized string `json:"subject_normalized"`
	State             string `json:"state"`
}

type payloadContact struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
	DisplayName  string `json:"display_name"`
}

type payloadContactMemory struct {
	ID        string   `json:"id"`
	ContactID string   `json:"contact_id"`
	ThreadID  string   `json:"thread_id"`
	Note      string   `json:"note"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

func mapMessage(message domain.Message) payloadMessage {
	return payloadMessage{
		ID:              message.ID,
		ThreadID:        message.ThreadID,
		Direction:       string(message.Direction),
		Subject:         message.Subject,
		TextBody:        message.TextBody,
		MessageIDHeader: message.MessageIDHeader,
		CreatedAt:       message.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
