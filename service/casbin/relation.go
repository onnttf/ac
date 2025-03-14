package casbin

import (
	"ac/bootstrap/database"
	"ac/dal"
	"ac/model"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

type Relation struct {
	System  string
	Group   string
	Subject string
}

func (o Relation) validate() error {
	if o.System == "" {
		return fmt.Errorf("system is empty")
	}
	if o.Group == "" {
		return fmt.Errorf("group is empty")
	}
	if o.Subject == "" {
		return fmt.Errorf("subject is empty")
	}
	return nil
}

func (o Relation) toRule() (*model.CasbinRule, error) {
	o.Group = strings.TrimSpace(o.Group)
	o.Subject = strings.TrimSpace(o.Subject)

	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("relation validate failed, err: %w", err)
	}
	return &model.CasbinRule{
		PType: model.PTypeGroup,
		V0:    o.System,
		V1:    o.Group,
		V2:    o.Subject,
	}, nil
}

func GetSubjectRoleList(ctx context.Context, system, subject string) ([]Relation, error) {
	recordList, err := dal.NewRepo[model.CasbinRule]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.CasbinRule{
			PType: model.PTypeGroup,
			V0:    system,
			V1:    subject,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subject permission, err: %w", err)
	}
	final := make([]Relation, 0, len(recordList))
	for _, v := range recordList {
		final = append(final, Relation{
			System:  v.V0,
			Group:   v.V1,
			Subject: v.V2,
		})
	}
	return final, nil
}
