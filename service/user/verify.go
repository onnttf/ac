package user

import (
	"context"
	"errors"
	"fmt"

	"ac/bootstrap/database"
	"ac/model"

	"github.com/onnttf/kit/dal"

	"gorm.io/gorm"
)

func Verify(ctx context.Context, code string) error {
	if code == "" {
		return errors.New("empty code")
	}

	userRepo := dal.NewRepo[model.TblSubject]()
	user, err := userRepo.QueryOne(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Select("type", "deleted").Where("code = ?", code)
	})
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}

	if user == nil {
		return errors.New("user not found")
	}

	if user.Type != model.SubjectTypeUser {
		return errors.New("not a user")
	}

	if user.Deleted == model.Deleted {
		return errors.New("user deleted")
	}

	return nil
}
