package postgres

import "time"

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
