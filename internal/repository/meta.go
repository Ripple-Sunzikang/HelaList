package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func GetMetaByPath(path string) (*model.Meta, error) {
	meta := model.Meta{Path: path}
	if err := bootstrap.Db.Where(meta).First(&meta).Error; err != nil {
		return nil, errors.Wrapf(err, "failed select meta")
	}
	return &meta, nil
}

func GetMetaById(id uuid.UUID) (*model.Meta, error) {
	var u model.Meta
	if err := bootstrap.Db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get old meta")
	}
	return &u, nil
}

func CreateMeta(u *model.Meta) error {
	return errors.WithStack(bootstrap.Db.Create(u).Error)
}

func UpdateMeta(u *model.Meta) error {
	return errors.WithStack(bootstrap.Db.Save(u).Error)
}

func GetMetas(pageIndex, pageSize int) (metas []model.Meta, count int64, err error) {
	metaDB := bootstrap.Db.Model(&model.Meta{})
	if err = metaDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get metas count")
	}
	if err = metaDB.Order(columnName("id")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&metas).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get find metas")
	}
	return metas, count, nil
}

func DeleteMetaById(id uuid.UUID) error {
	return errors.WithStack(bootstrap.Db.Delete(&model.Meta{}, id).Error)
}
