package op

import (
	"HelaList/internal/driver"
	"HelaList/internal/model"
	"context"
	"slices"
	"time"

	stdpath "path"

	"github.com/OpenListTeam/OpenList/v4/pkg/generic_sync"
	"github.com/OpenListTeam/OpenList/v4/pkg/singleflight"
	"github.com/OpenListTeam/go-cache"
	"github.com/pkg/errors"
)

/*
本着尽可能原创的精神，本项目已经在尽最大可能不要去直接剽窃他人成果了。
当然，一些成品第三方库是可以使用的。
但是，但是，OpenList的库我不想用，不然这和抄袭有什么区别？
但从实际来说，singleflight和go-cache都是基于其他人的库进行设计的。
而OpenList做的事情，就是把他们的所有函数转为了泛型。
所以你，要么把官方的singleflight和那个第三方的go-cache自己写一遍。
要么，你就顺从OpenList。毕竟，你要是自己魔改，无非也就是把他们改成泛型了。
*/

var listCache = cache.NewMemCache(cache.WithShards[[]model.Obj](64)) // 文件缓存
// singleflight用于将短时间内对多个统一key的并发请求进行合并，防止请求出现冲突
var listGroup singleflight.Group[[]model.Obj]

// 将挂载路径与原始路径合并
func Key(storage driver.Driver, path string) string {
	return stdpath.Join(storage.GetStorage().MountPath, path)
}

// 缓存替换算法
func updateCacheObj(storage driver.Driver, path string, oldObj model.Obj, newObj model.Obj) {
	key := Key(storage, path)
	objs, ok := listCache.Get(key) // 获取该路径对应缓存
	if ok {
		for i, obj := range objs {
			if obj.GetName() == newObj.GetName() {
				objs = slices.Delete(objs, i, i+1) // 清理缓存旧数据
				break
			}
		}
		for i, obj := range objs {
			if obj.GetName() == oldObj.GetName() {
				objs[i] = newObj
				break
			}
		}
		listCache.Set(key, objs, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
	}
}

// 删除缓存内容
func delCacheObj(storage driver.Driver, path string, obj model.Obj) {
	key := Key(storage, path)
	objs, ok := listCache.Get(key)
	if ok {
		for i, oldObj := range objs {
			if oldObj.GetName() == obj.GetName() {
				objs = append(objs[:i], objs[i+1:]...)
				break
			}
		}
		listCache.Set(key, objs, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
	}
}

// 清空缓存
func ClearCache(storage driver.Driver, path string) {
	objs, ok := listCache.Get(Key(storage, path))
	if ok {
		for _, obj := range objs {
			if obj.IsDir() {
				ClearCache(storage, stdpath.Join(path, obj.GetName()))
			}
		}
	}
	listCache.Del(Key(storage, path))
}

// 删除缓存内容
func DeleteCache(storage driver.Driver, path string) {
	listCache.Del(Key(storage, path))
}

// 函数式编程太好用了你们知道吗
var addSortDebounceMap generic_sync.MapOf[string, func(func())]

func addCacheObj(storage driver.Driver, path string, newObj model.Obj) {
	key := Key(storage, path)
	objs, ok := listCache.Get(key)
	if ok {
		for i, obj := range objs {
			if obj.GetName() == newObj.GetName() {
				objs[i] = newObj
				return
			}
		}

		// 文件/文件夹分离
		if len(objs) > 0 && objs[len(objs)-1].IsDir() == newObj.IsDir() {
			objs = append(objs, newObj)
		} else {
			objs = append([]model.Obj{newObj}, objs...)
		}

		listCache.Set(key, objs, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
	}
}

// 以上，缓存功能到此为止。几乎都是靠调库实现的。

// 列举storage中的文件，但不包含虚拟文件
func List(ctx context.Context, storage driver.Driver, path string, args model.ListArgs) ([]model.Obj, error) {
	key := Key(storage, path)
	if !args.Refresh {
		if files, ok := listCache.Get(key); ok {
			return files, nil
		}
	}
	dir, err := GetUnwrap(ctx, storage, path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get dir")
	}
	if !dir.IsDir() {
		return nil, errors.WithMessage(err, "not folder")
	}
	objs, err, _ := listGroup.Do(key, func() ([]model.Obj, error) {
		files, err := storage.List(ctx, dir, args)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list objs")
		}
		for _, f := range files {
			if s, ok := f.(model.SetPath); ok && f.GetPath() == "" && dir.GetPath() != "" {
				s.SetPath(stdpath.Join(dir.GetPath(), f.GetName()))
			}
		}
		model.WrapObjsName(files)

		go func(reqPath string, files []model.Obj) {
			HandleObjsUpdateHook(reqPath, files)
		}(stdpath.Join(storage.GetStorage().MountPath, path), files)

		// 诶我突然就很好奇，万一不sort会怎么样。
		// // sort objs
		// if storage.Config().LocalSort {
		// 	model.SortFiles(files, storage.GetStorage().OrderBy, storage.GetStorage().OrderDirection)
		// }
		// model.ExtractFolder(files, storage.GetStorage().ExtractFolder)

		// if !storage.Config().NoCache {
		// 	if len(files) > 0 {
		// 		log.Debugf("set cache: %s => %+v", key, files)
		// 		listCache.Set(key, files, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
		// 	} else {
		// 		log.Debugf("del cache: %s", key)
		// 		listCache.Del(key)
		// 	}
		// }
		return files, nil
	})
	return objs, err
}

func Get(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	if g, ok := storage.(driver.Getter); ok {
		obj, err := g.Get(ctx, path)
		if err == nil {
			return model.WrapObjName(obj), nil
		}
	}

	// is root folder
	if path == "/" {
		var rootObj model.Obj
		if getRooter, ok := storage.(driver.GetRooter); ok {
			obj, err := getRooter.GetRoot(ctx)
			if err != nil {
				return nil, errors.WithMessage(err, "failed get root obj")
			}
			rootObj = obj
		} else {
			switch r := storage.GetAddition().(type) {
			case driver.IRootId:
				rootObj = &model.Object{
					Id:           r.GetRootId(),
					Name:         "root",
					Size:         0,
					ModifiedTime: storage.GetStorage().ModifiedTime,
					IsFolder:     true,
				}
			case driver.IRootPath:
				rootObj = &model.Object{
					Path:         r.GetRootPath(),
					Name:         "root",
					Size:         0,
					ModifiedTime: storage.GetStorage().ModifiedTime,
					IsFolder:     true,
				}
			default:
				return nil, errors.Errorf("please implement IRootPath or IRootId or GetRooter method")
			}
		}
		if rootObj == nil {
			return nil, errors.Errorf("please implement IRootPath or IRootId or GetRooter method")
		}
		return &model.ObjWrapName{
			Name: "root",
			Obj:  rootObj,
		}, nil
	}

	// not root folder: try list parent and find by name
	dir, name := stdpath.Split(path)
	files, err := List(ctx, storage, dir, model.ListArgs{})
	if err != nil {
		return nil, errors.WithMessage(err, "failed get parent list")
	}
	for _, f := range files {
		if f.GetName() == name {
			return model.WrapObjName(f), nil
		}
	}
	return nil, errors.Errorf("object not found: %s", path)
}

func GetUnwrap(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	obj, err := Get(ctx, storage, path)
	if err != nil {
		return nil, err
	}
	return model.UnwrapObj(obj), err
}
