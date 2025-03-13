package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameResource = "resource"

// Resource mapped from table <resource>
type Resource struct {
	ID          int64          `gorm:"column:id;type:int;primaryKey;autoIncrement:true;comment:id" json:"id"`                                                                                                                                             // id
	SystemCode  string         `gorm:"column:system_code;type:varchar(50);not null;uniqueIndex:uk_system_code_code,priority:1;index:idx_system_code_name,priority:1;index:idx_system_code_parent_code,priority:1;comment:system_code" json:"system_code"` // system_code
	Code        string         `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_system_code_code,priority:2;comment:code" json:"code"`                                                                                                         // code
	Name        string         `gorm:"column:name;type:varchar(50);not null;index:idx_system_code_name,priority:2;comment:name" json:"name"`                                                                                                              // name
	Description *string        `gorm:"column:description;type:varchar(50);not null;comment:description" json:"description"`                                                                                                                               // description
	ParentCode  *string        `gorm:"column:parent_code;type:varchar(50);not null;index:idx_system_code_parent_code,priority:2;comment:parent_code" json:"parent_code"`                                                                                  // parent_code
	ModifiedBy  string         `gorm:"column:modified_by;type:varchar(50);not null;comment:modified_by" json:"modified_by"`                                                                                                                               // modified_by
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:created_at" json:"created_at"`                                                                                                           // created_at
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:updated_at" json:"updated_at"`                                                                                                           // updated_at
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;comment:deleted_at" json:"deleted_at"`                                                                                                                                              // deleted_at
}

// TableName Resource's table name
func (*Resource) TableName() string {
	return TableNameResource
}
