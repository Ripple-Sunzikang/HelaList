package repository

import (
	"fmt"

	"HelaList/configs"

	"gorm.io/gorm"
)

func columnName(name string) string {
	if configs.Conf.Database.Type == "postgres" {
		return fmt.Sprintf(`"%s"`, name)
	}
	return fmt.Sprintf("`%s`", name)
}

func addStorageOrder(db *gorm.DB) *gorm.DB {
	return db.Order(fmt.Sprintf("%s, %s", columnName("order"), columnName("id")))
}
