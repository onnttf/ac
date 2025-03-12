package role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/custom/code"
	"ac/dal"
	"ac/model"
	"context"
	"fmt"
	"gorm.io/gorm"
)

// GenerateCode creates a unique system code, trying up to maxAttempts times.
// Returns the generated code if successful, or an error if unable to generate a unique code.
// Considers even deleted records to ensure code uniqueness across all data.
func GenerateCode(ctx context.Context) (string, error) {
	const maxAttempts = 3

	for i := 0; i < maxAttempts; i++ {
		// Generate a new code
		tmpCode := code.GenerateRoleCode()

		// Check if the code exists (including deleted records)
		record, err := dal.NewRepo[model.Role]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
			return db.Unscoped().Where(model.Role{Code: tmpCode})
		})
		if err != nil {
			return "", fmt.Errorf("failed to query system, err: %w, code: %s", err, tmpCode)
		}

		// If no record found, return the generated code
		if record == nil {
			return tmpCode, nil
		}
	}

	// If we tried maxAttempts times and couldn't find a unique code
	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// Validate checks if a non-deleted system with the given code exists.
// Returns true if the system exists and is active, false otherwise.
func Validate(ctx context.Context, systemCode, code string) (bool, error) {
	if code == "" {
		return false, fmt.Errorf("code is empty")
	}

	// Only consider non-deleted records for validation
	record, err := dal.NewRepo[model.Role]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Role{SystemCode: systemCode, Code: code})
	})
	if err != nil {
		return false, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
	}

	if record == nil {
		logger.Infof(ctx, "no matching active system found, code: %s", code)
		return false, nil
	}

	return true, nil
}
