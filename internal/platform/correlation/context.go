package correlation

import "context"

type contextKey string

const correlationIDKey contextKey = "correlation_id"

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

func ID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(correlationIDKey).(string)
	if !ok || value == "" {
		return "", false
	}

	return value, true
}
