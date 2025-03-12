package code

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	PrefixLog      = "log"
	PrefixResource = "resource"
	PrefixRole     = "role"
	PrefixUser     = "user"
	PrefixSystem   = "system"
)

func GenerateLogCode() string {
	return generateCode(PrefixLog)
}

func GenerateResourceCode() string {
	return generateCode(PrefixResource)
}

func GenerateRoleCode() string {
	return generateCode(PrefixRole)
}

func GenerateUserCode() string {
	return generateCode(PrefixUser)
}

func GenerateSystemCode() string {
	return generateCode(PrefixSystem)
}

// generateCode generates a unique code with the given prefix.
func generateCode(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
