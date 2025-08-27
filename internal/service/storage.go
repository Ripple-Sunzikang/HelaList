package service

import (
	"HelaList/internal/model"
	"HelaList/internal/repository"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func CreateStorage(storage *model.Storage) error {
	if storage == nil {
		return errors.New("storage cannot be nil")
	}
	return repository.CreateStorage(storage)
}

func UpdateStorage(storage *model.Storage) error {
	if storage == nil {
		return errors.New("storage cannot be nil")
	}
	return repository.UpdateStorage(storage)
}

func DeleteStorageById(id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid storage id")
	}
	return repository.DeleteStorageById(id)
}

func GetStorages(pageIndex, pageSize int) ([]model.Storage, int64, error) {
	if pageIndex < 1 || pageSize < 1 {
		return nil, 0, errors.New("invalid pagination parameters")
	}
	return repository.GetStorages(pageIndex, pageSize)
}

func GetStorageById(id uuid.UUID) (*model.Storage, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid storage id")
	}
	return repository.GetStorageById(id)
}

func GetStorageByMountPath(mountPath string) (*model.Storage, error) {
	if mountPath == "" {
		return nil, errors.New("mount path cannot be empty")
	}
	return repository.GetStorageByMountPath(mountPath)
}

func GetEnabledStorages() ([]model.Storage, error) {
	return repository.GetEnabledStorages()
}
