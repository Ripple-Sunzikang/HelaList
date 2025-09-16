package op

import (
	stdpath "path"
	"time"

	"HelaList/internal/model"
	"HelaList/internal/service"

	"github.com/OpenListTeam/OpenList/v4/pkg/singleflight"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/go-cache"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var metaCache = cache.NewMemCache(cache.WithShards[*model.Meta](2))
var metaG singleflight.Group[*model.Meta]

func GetNearestMeta(path string) (*model.Meta, error) {
	return getNearestMeta(utils.FixAndCleanPath(path))
}
func getNearestMeta(path string) (*model.Meta, error) {
	meta, err := GetMetaByPath(path)
	if err == nil {
		return meta, nil
	}
	if errors.Cause(err) != gorm.ErrRecordNotFound {
		return nil, err
	}
	if path == "/" {
		return nil, gorm.ErrRecordNotFound
	}
	return getNearestMeta(stdpath.Dir(path))
}

func GetMetaByPath(path string) (*model.Meta, error) {
	return getMetaByPath(utils.FixAndCleanPath(path))
}
func getMetaByPath(path string) (*model.Meta, error) {
	meta, ok := metaCache.Get(path)
	if ok {
		if meta == nil {
			return meta, gorm.ErrRecordNotFound
		}
		return meta, nil
	}
	meta, err, _ := metaG.Do(path, func() (*model.Meta, error) {
		_meta, err := service.GetMetaByPath(path)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				metaCache.Set(path, nil)
				return nil, gorm.ErrRecordNotFound
			}
			return nil, err
		}
		metaCache.Set(path, _meta, cache.WithEx[*model.Meta](time.Hour))
		return _meta, nil
	})
	return meta, err
}

func DeleteMetaById(id uuid.UUID) error {
	old, err := service.GetMetaById(id)
	if err != nil {
		return err
	}
	metaCache.Del(old.Path)
	return service.DeleteMetaById(id)
}

func UpdateMeta(u *model.Meta) error {
	u.Path = utils.FixAndCleanPath(u.Path)
	old, err := service.GetMetaById(u.ID)
	if err != nil {
		return err
	}
	metaCache.Del(old.Path)
	return service.UpdateMeta(u)
}

func CreateMeta(u *model.Meta) error {
	u.Path = utils.FixAndCleanPath(u.Path)
	metaCache.Del(u.Path)
	return service.CreateMeta(u)
}

func GetMetaById(id uuid.UUID) (*model.Meta, error) {
	return service.GetMetaById(id)
}

func GetMetas(pageIndex, pageSize int) (metas []model.Meta, count int64, err error) {
	return service.GetMetas(pageIndex, pageSize)
}
