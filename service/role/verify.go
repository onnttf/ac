package role

import (
	"context"
	"errors"
	"fmt"

	"ac/bootstrap/database"
	"ac/model"

	"github.com/onnttf/kit/container"
	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

func BatchVerify(ctx context.Context, codes []string) (map[string]error, error) {
	if len(codes) == 0 {
		return nil, errors.New("empty codes")
	}

	codes = container.Deduplicate(codes)

	roleRepo := dal.NewRepo[model.TblSubject]()
	roles, err := roleRepo.Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Select("code", "type", "deleted").Where("code IN ?", codes)
	})
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}

	roleMap := make(map[string]*model.TblSubject, len(roles))
	for i := range roles {
		roleMap[roles[i].Code] = &roles[i]
	}

	result := make(map[string]error)
	for _, code := range codes {
		role, exists := roleMap[code]
		if !exists {
			result[code] = errors.New("role not found")
			continue
		}
		if role.Type != model.SubjectTypeRole {
			result[code] = errors.New("not a role")
			continue
		}
		if role.Deleted == model.Deleted {
			result[code] = errors.New("role deleted")
			continue
		}
	}

	return result, nil
}
