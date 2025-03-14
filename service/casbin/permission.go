package casbin

import (
	"ac/bootstrap/database"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Permission struct {
	System    string
	Subject   string
	Index     string
	Action    string
	BeginTime time.Time
	EndTime   time.Time
}

func (o Permission) validate() error {
	if o.System == "" {
		return fmt.Errorf("system is empty")
	}
	if o.Subject == "" {
		return fmt.Errorf("subject is empty")
	}
	if o.Index == "" {
		return fmt.Errorf("index is empty")
	}
	if o.Action == "" {
		return fmt.Errorf("action is empty")
	}
	if _, ok := action2Level[o.Action]; !ok {
		return fmt.Errorf("invalid action: %s", o.Action)
	}
	if o.BeginTime.IsZero() {
		return fmt.Errorf("begin time is not set")
	}
	if o.EndTime.IsZero() {
		return fmt.Errorf("end time is not set")
	}
	if !o.EndTime.After(o.BeginTime) {
		return fmt.Errorf("end time must be after begin time")
	}
	return nil
}

func (o Permission) toRule() (*model.CasbinRule, error) {
	o.Subject = strings.TrimSpace(o.Subject)
	o.Index = strings.TrimSpace(o.Index)
	o.Action = strings.TrimSpace(o.Action)

	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("permission validate failed, err: %w", err)
	}

	beginTime := util.FormatToDateTime(o.BeginTime)
	endTime := util.FormatToDateTime(o.EndTime)
	return &model.CasbinRule{
		PType: model.PTypePolicy,
		V0:    o.System,
		V1:    o.Subject,
		V2:    o.Index,
		V3:    &o.Action,
		V4:    &beginTime,
		V5:    &endTime,
	}, nil
}

func GetSubjectPermissionList(ctx context.Context, system, subject string) ([]Permission, error) {
	recordList, err := dal.NewRepo[model.CasbinRule]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.CasbinRule{
			PType: model.PTypePolicy,
			V0:    system,
			V1:    subject,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subject permission, err: %w", err)
	}
	final := make([]Permission, 0, len(recordList))
	for _, record := range recordList {
		beginTime, err := util.ParseFromDateTime(*record.V4)
		if err != nil {
			return nil, fmt.Errorf("parse begin time failed, err: %w", err)
		}
		endTime, err := util.ParseFromDateTime(*record.V5)
		if err != nil {
			return nil, fmt.Errorf("parse end time failed, err: %w", err)
		}
		final = append(final, Permission{
			System:    record.V0,
			Subject:   record.V1,
			Index:     record.V2,
			Action:    *record.V3,
			BeginTime: beginTime,
			EndTime:   endTime,
		})
	}
	return final, nil
}
