package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32          `gorm:"primarykey;type:int" json:"id"`
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"` // 自定义字段名
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"` // gorm 逻辑删除字段
	IsDeleted bool           `gorm:"column:is_deleted" json:"-"`
}

// 自定义gorm数据类型
type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}
