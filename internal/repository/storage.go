package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func CreateStorage(storage *model.Storage) error {
	result := bootstrap.Db.Create(storage)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

func UpdateStorage(storage *model.Storage) error {
	result := bootstrap.Db.Save(storage)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

func DeleteStorageById(id uuid.UUID) error {
	if err := bootstrap.Db.Delete(&model.Storage{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete mount: %w", err)
	}
	return nil
}

func GetStorageById(id uuid.UUID) (*model.Storage, error) {
	var storage model.Storage
	storage.Id = id
	if err := bootstrap.Db.First(&storage).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &storage, nil
}

func GetStorageByMountPath(mountPath string) (*model.Storage, error) {
	var storage model.Storage
	if err := bootstrap.Db.Where("mounnt_path = ?", mountPath).First(&storage).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &storage, nil
}

/*
// 涉及disable,但是目前来说不是必须功能
func GetEnabledStorages() ([]model.Storage, error) {

}
*/
