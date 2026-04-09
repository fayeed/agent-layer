package parser

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/agentlayer/agentlayer/internal/core"
)

type RawMessageReader interface {
	Get(ctx context.Context, objectKey string) ([]byte, error)
}

type Parser struct {
	reader RawMessageReader
}

func New(reader RawMessageReader) Parser {
	return Parser{reader: reader}
}

func (p Parser) Parse(ctx context.Context, message core.StoredInboundMessage) (core.ParsedMessage, error) {
	raw, err := p.reader.Get(ctx, message.Receipt.RawMessageObjectKey)
	if err != nil {
		return core.ParsedMessage{}, err
	}

	m, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return core.ParsedMessage{}, err
	}

	parsed := core.ParsedMessage{
		MessageIDHeader:   m.Header.Get("Message-ID"),
		InReplyTo:         m.Header.Get("In-Reply-To"),
		References:        strings.Fields(m.Header.Get("References")),
		Subject:           decodeHeader(m.Header.Get("Subject")),
		SubjectNormalized: normalizeSubject(decodeHeader(m.Header.Get("Subject"))),
		From:              parseSingleAddress(m.Header.Get("From")),
		ReplyTo:           parseAddressList(m.Header.Get("Reply-To")),
		To:                parseAddressList(m.Header.Get("To")),
		CC:                parseAddressList(m.Header.Get("Cc")),
		RawHeaders:        map[string][]string(m.Header),
	}

	contentType := m.Header.Get("Content-Type")
	if err := populateBodiesAndAttachments(&parsed, contentType, m.Header.Get("Content-Transfer-Encoding"), m.Body); err != nil {
		return core.ParsedMessage{}, err
	}

	return parsed, nil
}

func populateBodiesAndAttachments(parsed *core.ParsedMessage, contentType, transferEncoding string, body io.Reader) error {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		payload, readErr := io.ReadAll(body)
		if readErr != nil {
			return readErr
		}
		parsed.TextBody = string(payload)
		return nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		reader := multipart.NewReader(body, params["boundary"])
		for {
			part, nextErr := reader.NextPart()
			if nextErr == io.EOF {
				return nil
			}
			if nextErr != nil {
				return nextErr
			}

			partContentType := part.Header.Get("Content-Type")
			disposition := part.Header.Get("Content-Disposition")
			if strings.HasPrefix(strings.ToLower(disposition), "attachment") {
				payload, readErr := readDecodedBody(part, part.Header.Get("Content-Transfer-Encoding"))
				_ = part.Close()
				if readErr != nil {
					return readErr
				}

				parsed.Attachments = append(parsed.Attachments, core.ParsedAttachment{
					FileName:           part.FileName(),
					ContentType:        mediaTypeFromHeader(partContentType),
					SizeBytes:          int64(len(payload)),
					ContentID:          part.Header.Get("Content-ID"),
					ContentDisposition: disposition,
				})
				continue
			}

			if err := populateBodiesAndAttachments(parsed, partContentType, part.Header.Get("Content-Transfer-Encoding"), part); err != nil {
				_ = part.Close()
				return err
			}

			_ = part.Close()
		}
	}

	payload, err := readDecodedBody(body, transferEncoding)
	if err != nil {
		return err
	}

	switch mediaType {
	case "text/plain":
		if parsed.TextBody == "" {
			parsed.TextBody = string(payload)
		}
	case "text/html":
		if parsed.HTMLBody == "" {
			parsed.HTMLBody = string(payload)
		}
	}

	return nil
}

func parseSingleAddress(value string) core.ParsedAddress {
	addresses := parseAddressList(value)
	if len(addresses) == 0 {
		return core.ParsedAddress{}
	}
	return addresses[0]
}

func parseAddressList(value string) []core.ParsedAddress {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	addresses, err := mail.ParseAddressList(value)
	if err != nil {
		return nil
	}

	parsed := make([]core.ParsedAddress, 0, len(addresses))
	for _, address := range addresses {
		parsed = append(parsed, core.ParsedAddress{
			Email:       address.Address,
			DisplayName: decodeHeader(address.Name),
		})
	}
	return parsed
}

func decodeHeader(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	decoded, err := new(mime.WordDecoder).DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}

func readDecodedBody(body io.Reader, transferEncoding string) ([]byte, error) {
	reader := body
	switch strings.ToLower(strings.TrimSpace(transferEncoding)) {
	case "base64":
		reader = base64.NewDecoder(base64.StdEncoding, body)
	case "quoted-printable":
		reader = quotedprintable.NewReader(body)
	}

	return io.ReadAll(reader)
}

func normalizeSubject(subject string) string {
	normalized := strings.TrimSpace(subject)
	for {
		lower := strings.ToLower(normalized)
		switch {
		case strings.HasPrefix(lower, "re:"):
			normalized = strings.TrimSpace(normalized[3:])
		case strings.HasPrefix(lower, "fwd:"):
			normalized = strings.TrimSpace(normalized[4:])
		case strings.HasPrefix(lower, "fw:"):
			normalized = strings.TrimSpace(normalized[3:])
		default:
			return strings.ToLower(normalized)
		}
	}
}

func mediaTypeFromHeader(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return contentType
	}
	return mediaType
}
