package repository

import (
	"fmt"

	"gorm.io/gorm"
)

func columnName(name string) string {
	// 直接返回 PostgreSQL 的格式，不再需要 if 判断
	return fmt.Sprintf(`"%s"`, name)
}

func addStorageOrder(db *gorm.DB) *gorm.DB {
	return db.Order(fmt.Sprintf("%s, %s", columnName("order"), columnName("id")))
}
