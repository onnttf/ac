package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameSystem = "system"

// System mapped from table <system>
type System struct {
	ID          int64          `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true;comment:id" json:"id"`                          // id
	Code        string         `gorm:"column:code;type:varchar(50);not null;uniqueIndex:uk_code,priority:1;comment:code" json:"code"`           // code
	Name        string         `gorm:"column:name;type:varchar(50);not null;index:idx_name,priority:1;comment:name" json:"name"`                // name
	Description string         `gorm:"column:description;type:varchar(50);not null;comment:description" json:"description"`                     // description
	ModifiedBy  string         `gorm:"column:modified_by;type:varchar(50);not null;comment:modified_by" json:"modified_by"`                     // modified_by
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:created_at" json:"created_at"` // created_at
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:updated_at" json:"updated_at"` // updated_at
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;comment:deleted_at" json:"deleted_at"`                                    // deleted_at
}

// TableName System's table name
func (*System) TableName() string {
	return TableNameSystem
}
