package casbin

import (
	"ac/model"
	"fmt"
	"strings"
)

type Relation struct {
	Group   string
	Subject string
}

func (o Relation) validate() error {
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
		V0:    o.Group,
		V1:    o.Subject,
	}, nil
}
