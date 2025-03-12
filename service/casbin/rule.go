package casbin

import (
	"ac/bootstrap/database"
	"ac/custom/code"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"context"
	"fmt"
	"gorm.io/gorm"
)

var action2Level = map[string]int{
	"view":     1,
	"download": 2,
	"edit":     3,
	"manage":   4,
}

type RuleProcessor interface {
	validate() error
	toRule() (*model.CasbinRule, error)
}

func Add(ctx context.Context, ruleList []RuleProcessor, modifiedBy string) error {
	now := util.UTCNow()

	ruleListToAdd := make([]*model.CasbinRule, 0, len(ruleList))
	logList := make([]*model.CasbinRuleLog, 0, len(ruleList))
	tmpCode := code.GenerateLogCode()
	for _, v := range ruleList {
		rule, err := v.toRule()
		if err != nil {
			return fmt.Errorf("failed to convert to rule, err: %w", err)
		}
		ruleListToAdd = append(ruleListToAdd, rule)
		logList = append(logList, &model.CasbinRuleLog{
			LogCode:    tmpCode,
			Operate:    model.OperateAdd,
			PType:      rule.PType,
			V0:         rule.V0,
			V1:         rule.V1,
			V2:         rule.V2,
			V3:         rule.V3,
			V4:         rule.V4,
			V5:         rule.V5,
			ModifiedBy: modifiedBy,
			CreatedAt:  now,
		})
	}

	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := dal.NewRepo[model.CasbinRule]().BatchInsert(ctx, tx, ruleListToAdd, 10); err != nil {
			return fmt.Errorf("failed to insert casbin_rule, err: %w", err)
		}
		if err := dal.NewRepo[model.CasbinRuleLog]().BatchInsert(ctx, tx, logList, 10); err != nil {
			return fmt.Errorf("failed to insert casbin_rule_log, err: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to commit add rule, err: %w", err)
	}
	return nil
}

func Delete(ctx context.Context, ruleList []RuleProcessor, modifiedBy string) error {
	now := util.UTCNow()

	ruleListToDelete := make([]*model.CasbinRule, 0, len(ruleList))
	logList := make([]*model.CasbinRuleLog, 0, len(ruleList))
	tmpCode := code.GenerateLogCode()
	for _, v := range ruleList {
		rule, err := v.toRule()
		if err != nil {
			return fmt.Errorf("failed to convert to rule, err: %w", err)
		}
		ruleListToDelete = append(ruleListToDelete, rule)
		logList = append(logList, &model.CasbinRuleLog{
			LogCode:    tmpCode,
			Operate:    model.OperateDelete,
			PType:      rule.PType,
			V0:         rule.V0,
			V1:         rule.V1,
			V2:         rule.V2,
			V3:         rule.V3,
			V4:         rule.V4,
			V5:         rule.V5,
			ModifiedBy: modifiedBy,
			CreatedAt:  now,
		})
	}

	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range ruleListToDelete {
			if err := dal.NewRepo[model.CasbinRule]().Delete(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(v).Delete(v)
			}); err != nil {
				return fmt.Errorf("failed to insert casbin_rule, err: %w", err)
			}
		}
		if err := dal.NewRepo[model.CasbinRule]().BatchInsert(ctx, tx, ruleListToDelete, 10); err != nil {
			return fmt.Errorf("failed to insert casbin_rule, err: %w", err)
		}
		if err := dal.NewRepo[model.CasbinRuleLog]().BatchInsert(ctx, tx, logList, 10); err != nil {
			return fmt.Errorf("failed to insert casbin_rule_log, err: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to commit delete rule, err: %w", err)
	}
	return nil
}

func Set(ctx context.Context, ruleList []RuleProcessor, modifiedBy string) error {
	now := util.UTCNow()

	ruleListToSet := make([]*model.CasbinRule, 0, len(ruleList))
	for _, v := range ruleList {
		rule, err := v.toRule()
		if err != nil {
			return fmt.Errorf("failed to convert to rule, err: %w", err)
		}
		ruleListToSet = append(ruleListToSet, rule)
	}

	tmpCode := code.GenerateLogCode()

	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range ruleListToSet {
			record, err := dal.NewRepo[model.CasbinRule]().Query(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(&model.CasbinRule{
					PType: v.PType,
					V0:    v.V0,
					V1:    v.V1,
				})
			})
			if err != nil {
				return fmt.Errorf("failed to query rule, err: %w", err)
			}
			if record != nil {
				if err := dal.NewRepo[model.CasbinRule]().Delete(ctx, tx, func(db *gorm.DB) *gorm.DB {
					return db.Where(v)
				}); err != nil {
					return fmt.Errorf("failed to delete rule, err: %w", err)
				}
				if err := dal.NewRepo[model.CasbinRuleLog]().Insert(ctx, tx, &model.CasbinRuleLog{
					LogCode:    tmpCode,
					Operate:    model.OperateDelete,
					PType:      v.PType,
					V0:         v.V0,
					V1:         v.V1,
					V2:         v.V2,
					V3:         v.V3,
					V4:         v.V4,
					V5:         v.V5,
					ModifiedBy: modifiedBy,
					CreatedAt:  now,
				}); err != nil {
					return fmt.Errorf("failed to insert casbin_rule_log, err: %w", err)
				}
			}
			if err := dal.NewRepo[model.CasbinRule]().Insert(ctx, tx, v); err != nil {
				return fmt.Errorf("failed to insert casbin_rule, err: %w", err)
			}
			if err := dal.NewRepo[model.CasbinRuleLog]().Insert(ctx, tx, &model.CasbinRuleLog{
				LogCode:    tmpCode,
				Operate:    model.OperateAdd,
				PType:      v.PType,
				V0:         v.V0,
				V1:         v.V1,
				V2:         v.V2,
				V3:         v.V3,
				V4:         v.V4,
				V5:         v.V5,
				ModifiedBy: modifiedBy,
				CreatedAt:  now,
			}); err != nil {
				return fmt.Errorf("failed to insert casbin_rule_log, err: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to commit set rule, err: %w", err)
	}

	return nil
}
