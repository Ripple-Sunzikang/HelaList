package fs

import (
	"HelaList/internal/model"
	"context"

	"HelaList/configs"
	"HelaList/internal/op"

	"github.com/pkg/errors"
	"google.golang.org/appengine/log"
)

// List files
func list(ctx context.Context, path string, args *ListArgs) ([]model.Obj, error) {
	meta, _ := ctx.Value(configs.MetaKey).(*model.Meta)
	user, _ := ctx.Value(configs.UserKey).(*model.User)
	virtualFiles := op.GetStorageVirtualFilesByPath(path)
	storage, actualPath, err := op.GetStorageAndActualPath(path)
	if err != nil && len(virtualFiles) == 0 {
		return nil, errors.WithMessage(err, "failed get storage")
	}

	var _objs []model.Obj
	if storage != nil {
		_objs, err = op.List(ctx, storage, actualPath, model.ListArgs{
			ReqPath: path,
			Refresh: args.Refresh,
		})
		if err != nil {
			if !args.NoLog {
				log.Errorf("fs/list: %+v", err)
			}
			if len(virtualFiles) == 0 {
				return nil, errors.WithMessage(err, "failed get objs")
			}
		}
	}

	om := model.NewObjMerge()
	if whetherHide(user, meta, path) {
		om.InitHideReg(meta.Hide)
	}
	objs := om.Merge(_objs, virtualFiles...)
	return objs, nil
}
