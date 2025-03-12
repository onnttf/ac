package model

import (
	"time"
)

const OperateAdd = 1
const OperateDelete = 2

const TableNameCasbinRuleLog = "casbin_rule_log"

// CasbinRuleLog mapped from table <casbin_rule_log>
type CasbinRuleLog struct {
	ID         int64     `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true;comment:id" json:"id"`                          // id
	LogCode    string    `gorm:"column:log_code;type:varchar(50);not null;comment:log_code" json:"log_code"`                              // log_code
	Operate    int       `gorm:"column:operate;type:tinyint;not null;comment:operate, 1. add, 2. delete" json:"operate"`                  // operate, 1. add, 2. delete
	PType      string    `gorm:"column:ptype;type:varchar(10);not null;index:idx_ptype_v0_v1,priority:1;comment:ptype" json:"ptype"`      // ptype
	V0         string    `gorm:"column:v0;type:varchar(255);not null;index:idx_ptype_v0_v1,priority:2;comment:v0" json:"v0"`              // v0
	V1         string    `gorm:"column:v1;type:varchar(255);not null;index:idx_ptype_v0_v1,priority:3;comment:v1" json:"v1"`              // v1
	V2         string    `gorm:"column:v2;type:varchar(255);not null;comment:v2" json:"v2"`                                               // v2
	V3         string    `gorm:"column:v3;type:varchar(255);not null;comment:v3" json:"v3"`                                               // v3
	V4         string    `gorm:"column:v4;type:varchar(255);not null;comment:v4" json:"v4"`                                               // v4
	V5         string    `gorm:"column:v5;type:varchar(255);not null;comment:v5" json:"v5"`                                               // v5
	ModifiedBy string    `gorm:"column:modified_by;type:varchar(50);not null;comment:modified_by" json:"modified_by"`                     // modified_by
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:created_at" json:"created_at"` // created_at
}

// TableName CasbinRuleLog's table name
func (*CasbinRuleLog) TableName() string {
	return TableNameCasbinRuleLog
}
