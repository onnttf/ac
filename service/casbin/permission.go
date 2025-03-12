package casbin

import (
	"ac/model"
	"fmt"
	"strings"
	"time"
)

type Permission struct {
	Subject   string
	Resource  string
	Action    string
	BeginTime time.Time
	EndTime   time.Time
}

func (o Permission) validate() error {
	if o.Subject == "" {
		return fmt.Errorf("subject is empty")
	}
	if o.Resource == "" {
		return fmt.Errorf("resource is empty")
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
	o.Resource = strings.TrimSpace(o.Resource)
	o.Action = strings.TrimSpace(o.Action)

	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("permission validate failed, err: %w", err)
	}

	return &model.CasbinRule{
		PType: model.PTypePolicy,
		V0:    o.Subject,
		V1:    o.Resource,
		V2:    o.Action,
		V3:    o.BeginTime.Format(time.DateTime),
		V4:    o.EndTime.Format(time.DateTime),
	}, nil
}
