package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"fmt"

	"github.com/google/uuid"
)

func CreateMount(mount *model.Mount) error {
	result := bootstrap.Db.Create(mount)
	if result.Error != nil {
		return fmt.Errorf("failed to create mount: %w", result.Error)
	}
	return nil
}

func UpdateMount(mount *model.Mount) error {
	result := bootstrap.Db.Save(mount)
	if result.Error != nil {
		return fmt.Errorf("failed to update mount: %w", result.Error)
	}
	return nil
}

func DeleteMountById(id uuid.UUID) error {
	if err := bootstrap.Db.Delete(&model.Mount{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete mount: %w", err)
	}
	return nil
}
