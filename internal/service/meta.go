package service

import (
	"HelaList/internal/model"
	"HelaList/internal/repository"

	"github.com/pkg/errors"
)

func CreateMeta(meta *model.Meta) error {
	if meta == nil {
		return errors.New("meta cannot be nil")
	}
	if meta.Path == "" {
		return errors.New("meta path cannot be empty")
	}
	return repository.CreateMeta(meta)
}

func UpdateMeta(meta *model.Meta) error {
	if meta == nil {
		return errors.New("meta cannot be nil")
	}
	if meta.Path == "" {
		return errors.New("meta path cannot be empty")
	}
	return repository.UpdateMeta(meta)
}

func DeleteMetaById(id uint) error {
	if id == 0 {
		return errors.New("invalid meta id")
	}
	return repository.DeleteMetaById(id)
}

func GetMetas(pageIndex, pageSize int) ([]model.Meta, int64, error) {
	if pageIndex < 1 || pageSize < 1 {
		return nil, 0, errors.New("invalid pagination parameters")
	}
	return repository.GetMetas(pageIndex, pageSize)
}

func GetMetaById(id uint) (*model.Meta, error) {
	if id == 0 {
		return nil, errors.New("invalid meta id")
	}
	return repository.GetMetaById(id)
}

func GetMetaByPath(path string) (*model.Meta, error) {
	if path == "" {
		return nil, errors.New("meta path cannot be empty")
	}
	return repository.GetMetaByPath(path)
}
