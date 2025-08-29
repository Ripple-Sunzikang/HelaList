package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func CreateStorage(storage *model.Storage) error {
	return errors.WithStack(bootstrap.Db.Create(storage).Error)
}

func UpdateStorage(storage *model.Storage) error {
	return errors.WithStack(bootstrap.Db.Save(storage).Error)
}

func DeleteStorageById(id uuid.UUID) error {
	return errors.WithStack(bootstrap.Db.Delete(&model.Storage{}, id).Error)
}

func GetStorages(pageIndex, pageSize int) ([]model.Storage, int64, error) {
	storageDB := bootstrap.Db.Model(&model.Storage{})
	var count int64
	if err := storageDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get storages count")
	}
	var storages []model.Storage
	if err := addStorageOrder(storageDB).Order(columnName("order")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&storages).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return storages, count, nil
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
	if err := bootstrap.Db.Where("mount_path = ?", mountPath).First(&storage).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &storage, nil
}

func GetEnabledStorages() ([]model.Storage, error) {
	var storages []model.Storage
	err := addStorageOrder(bootstrap.Db).Where(fmt.Sprintf("%s = ?", columnName("disabled")), false).Find(&storages).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return storages, nil
}

/*
// 涉及disable,但是目前来说不是必须功能
func GetEnabledStorages() ([]model.Storage, error) {

}
*/
