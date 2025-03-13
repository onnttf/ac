package model

const PTypePolicy = "p"
const PTypeGroup = "g"

const TableNameCasbinRule = "casbin_rule"

// CasbinRule mapped from table <casbin_rule>
type CasbinRule struct {
	ID    int64   `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true;comment:id" json:"id"`                          // id
	PType string  `gorm:"column:ptype;type:varchar(10);not null;uniqueIndex:uk_ptype_v0_v1,priority:1;comment:ptype" json:"ptype"` // ptype
	V0    string  `gorm:"column:v0;type:varchar(255);not null;uniqueIndex:uk_ptype_v0_v1,priority:2;comment:v0" json:"v0"`         // v0
	V1    string  `gorm:"column:v1;type:varchar(255);not null;uniqueIndex:uk_ptype_v0_v1,priority:3;comment:v1" json:"v1"`         // v1
	V2    *string `gorm:"column:v2;type:varchar(255);not null;comment:v2" json:"v2"`                                               // v2
	V3    *string `gorm:"column:v3;type:varchar(255);not null;comment:v3" json:"v3"`                                               // v3
	V4    *string `gorm:"column:v4;type:varchar(255);not null;comment:v4" json:"v4"`                                               // v4
	V5    *string `gorm:"column:v5;type:varchar(255);not null;comment:v5" json:"v5"`                                               // v5
}

// TableName CasbinRule's table name
func (*CasbinRule) TableName() string {
	return TableNameCasbinRule
}
