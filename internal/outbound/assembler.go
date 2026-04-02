package outbound

import (
	"errors"
	"fmt"
	"strings"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type MessageIDGenerator func() string

type Assembler struct {
	generateMessageID MessageIDGenerator
}

type ReplyAssemblyInput struct {
	Organization   domain.Organization
	Agent          domain.Agent
	Inbox          domain.Inbox
	Thread         domain.Thread
	ReplyToMessage domain.Message
	Contact        domain.Contact
	BodyText       string
}

type ReplyMetadata struct {
	MessageIDHeader string
	Subject         string
	InReplyTo       string
	References      []string
}

func NewAssembler(generateMessageID MessageIDGenerator) Assembler {
	if generateMessageID == nil {
		generateMessageID = func() string {
			return "<generated@agentlayer.local>"
		}
	}

	return Assembler{
		generateMessageID: generateMessageID,
	}
}

func (a Assembler) AssembleReply(input ReplyAssemblyInput) (string, ReplyMetadata, error) {
	if input.Inbox.EmailAddress == "" {
		return "", ReplyMetadata{}, errors.New("inbox email address is required")
	}
	if input.Contact.EmailAddress == "" {
		return "", ReplyMetadata{}, errors.New("contact email address is required")
	}
	if input.ReplyToMessage.MessageIDHeader == "" {
		return "", ReplyMetadata{}, errors.New("reply-to message id header is required")
	}

	messageID := a.generateMessageID()
	subject := replySubject(input.ReplyToMessage.Subject)
	references := append(append([]string{}, input.ReplyToMessage.References...), input.ReplyToMessage.MessageIDHeader)
	from := formatAddress(input.Inbox.DisplayName, input.Inbox.EmailAddress)
	to := formatAddress(input.Contact.DisplayName, input.Contact.EmailAddress)

	raw := strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Message-ID: %s", messageID),
		fmt.Sprintf("In-Reply-To: %s", input.ReplyToMessage.MessageIDHeader),
		fmt.Sprintf("References: %s", strings.Join(references, " ")),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		input.BodyText,
	}, "\r\n")

	return raw, ReplyMetadata{
		MessageIDHeader: messageID,
		Subject:         subject,
		InReplyTo:       input.ReplyToMessage.MessageIDHeader,
		References:      references,
	}, nil
}

func replySubject(subject string) string {
	if strings.HasPrefix(strings.ToLower(subject), "re:") {
		return subject
	}

	if strings.TrimSpace(subject) == "" {
		return "Re:"
	}

	return "Re: " + subject
}

func formatAddress(name, email string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
