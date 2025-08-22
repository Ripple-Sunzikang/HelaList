package database

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"fmt"
)

func CreateMount(mount *model.Mount) error {

	result := bootstrap.Db.Create(mount)

	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}

	return nil
}
