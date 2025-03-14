package util

import (
	"fmt"
	"time"
)

// UTCNow returns the current time in UTC.
func UTCNow() time.Time {
	return time.Now().UTC()
}

func FormatToDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseFromDateTime(dateTime string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, dateTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime: %w", err)
	}
	return t, nil
}
