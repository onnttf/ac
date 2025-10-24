package util

import (
	"github.com/google/uuid"
)

func GenerateCode() string {
	u := uuid.New()
	return u.String()
}
