package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，所有业务模型均继承。
// 仅含时间戳与逻辑删除字段；CreateBy/UpdateBy 等审计字段按设计不写入（见路线图阶段八的取舍）。
type BaseModel struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 逻辑删除
}
