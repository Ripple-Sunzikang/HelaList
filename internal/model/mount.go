package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 后端实际存储的配置和元数据

type Mount struct {
	Id        uuid.UUID `gorm:"primaryKey" json:"id"`
	MountPath string    `gorm:"unique" json:"mount_path" binding:"required"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	Driver    string    `json:"driver"` // 采用什么驱动挂载，项目初期使用WebDAV
}

// 用于指定模型对应的数据库表名，模型的属性也会自动转化为列。默认为蛇形复数形式。
func (Mount) TableName() string {
	return "mounts"
}

func (m *Mount) BeforeCreate(tx *gorm.DB) (err error) {
	if m.Id == uuid.Nil {
		newUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		m.Id = newUUID
	}
	return nil
}
