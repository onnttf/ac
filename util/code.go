package util

import (
	"github.com/google/uuid"
)

// GenerateCode returns a new UUID string.
func GenerateCode() string {
	u := uuid.New()
	return u.String()
}
