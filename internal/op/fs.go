package op

import (
	"HelaList/configs"
	"HelaList/internal/driver"
	"HelaList/internal/model"
	"context"
	stderrors "errors"
	"fmt"
	"slices"
	"strings"
	"time"

	stdpath "path"

	"github.com/OpenListTeam/OpenList/v4/pkg/generic_sync"
	"github.com/OpenListTeam/OpenList/v4/pkg/singleflight"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/go-cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/*
本着尽可能原创的精神，本项目已经在尽最大可能不要去直接挪用他人的成品了。
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

// 缓存部分

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

// 下面部分主要负责实现fs层的底层设计

// Wrap是显示名映射功能，GetUnwrap负责解开映射
func GetUnwrap(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	obj, err := Get(ctx, storage, path)
	if err != nil {
		return nil, err
	}
	return model.UnwrapObj(obj), err
}

func Get(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	path = utils.FixAndCleanPath(path)
	if g, ok := storage.(driver.Getter); ok {
		obj, err := g.Get(ctx, path)
		if err == nil {
			return model.WrapObjName(obj), nil
		}
	}

	// 对路径是根目录的情况
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

	// 路径不是根目录的情况
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

// 列举storage中的文件，但不包含虚拟文件
func List(ctx context.Context, storage driver.Driver, path string, args model.ListArgs) ([]model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	path = utils.FixAndCleanPath(path)
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
			if s, ok := f.(model.SetPath); ok {
				s.SetPath(stdpath.Join(args.ReqPath, f.GetName()))
			}
		}
		model.WrapObjsName(files)

		go func(reqPath string, files []model.Obj) {
			HandleObjsUpdateHook(reqPath, files)
		}(utils.GetFullPath(storage.GetStorage().MountPath, path), files)

		// 诶我突然就很好奇，万一不sort会怎么样。
		// // sort objs
		// if storage.Config().LocalSort {
		// 	model.SortFiles(files, storage.GetStorage().OrderBy, storage.GetStorage().OrderDirection)
		// }
		// model.ExtractFolder(files, storage.GetStorage().ExtractFolder)

		if !storage.Config().NoCache {
			if len(files) > 0 {
				// log.Debugf("set cache: %s => %+v", key, files)
				listCache.Set(key, files, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
			} else {
				// log.Debugf("del cache: %s", key)
				listCache.Del(key)
			}
		}
		return files, nil
	})
	return objs, err
}

var mkdirG singleflight.Group[interface{}]

func MakeDir(ctx context.Context, storage driver.Driver, path string, lazyCache ...bool) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	path = utils.FixAndCleanPath(path)
	key := Key(storage, path)
	_, err, _ := mkdirG.Do(key, func() (interface{}, error) {
		// check if dir exists
		f, err := GetUnwrap(ctx, storage, path)
		if err != nil {
			if strings.Contains(err.Error(), "object not found") {
				parentPath, dirName := stdpath.Split(path)
				err = MakeDir(ctx, storage, parentPath)
				if err != nil {
					return nil, errors.WithMessagef(err, "failed to make parent dir [%s]", parentPath)
				}
				parentDir, err := GetUnwrap(ctx, storage, parentPath)
				if err != nil {
					return nil, errors.WithMessagef(err, "failed to get parent dir [%s]", parentPath)
				}

				switch s := storage.(type) {
				case driver.MkdirResult:
					var newObj model.Obj
					newObj, err = s.MakeDir(ctx, parentDir, dirName)
					if err == nil {
						if newObj != nil {
							addCacheObj(storage, parentPath, model.WrapObjName(newObj))
						} else if !utils.IsBool(lazyCache...) {
							DeleteCache(storage, parentPath)
						}
					}
				case driver.Mkdir:
					err = s.MakeDir(ctx, parentDir, dirName)
					if err == nil && !utils.IsBool(lazyCache...) {
						DeleteCache(storage, parentPath)
					}
				default:
					return nil, errors.New("not implemented")
				}
				return nil, errors.WithStack(err)
			}
			return nil, errors.WithMessage(err, "failed to check if dir exists")
		}
		if f.IsDir() {
			return nil, nil
		}
		return nil, errors.New("file exists")
	})
	return err
}

func Move(ctx context.Context, storage driver.Driver, srcPath, dstDirPath string, lazyCache ...bool) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcPath = utils.FixAndCleanPath(srcPath)
	dstDirPath = utils.FixAndCleanPath(dstDirPath)
	srcRawObj, err := Get(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	srcObj := model.UnwrapObj(srcRawObj)
	dstDir, err := GetUnwrap(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get dst dir")
	}
	srcDirPath := stdpath.Dir(srcPath)

	switch s := storage.(type) {
	case driver.MoveResult:
		var newObj model.Obj
		newObj, err = s.Move(ctx, srcObj, dstDir)
		if err == nil {
			delCacheObj(storage, srcDirPath, srcRawObj)
			if newObj != nil {
				addCacheObj(storage, dstDirPath, model.WrapObjName(newObj))
			} else if !utils.IsBool(lazyCache...) {
				DeleteCache(storage, dstDirPath)
			}
		}
	case driver.Move:
		err = s.Move(ctx, srcObj, dstDir)
		if err == nil {
			delCacheObj(storage, srcDirPath, srcRawObj)
			if !utils.IsBool(lazyCache...) {
				DeleteCache(storage, dstDirPath)
			}
		}
	default:
		return errors.New("not implemented")
	}
	return errors.WithStack(err)
}

func Rename(ctx context.Context, storage driver.Driver, srcPath, dstName string, lazyCache ...bool) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcPath = utils.FixAndCleanPath(srcPath)
	srcRawObj, err := Get(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	srcObj := model.UnwrapObj(srcRawObj)
	srcDirPath := stdpath.Dir(srcPath)

	switch s := storage.(type) {
	case driver.RenameResult:
		var newObj model.Obj
		newObj, err = s.Rename(ctx, srcObj, dstName)
		if err == nil {
			if newObj != nil {
				updateCacheObj(storage, srcDirPath, srcRawObj, model.WrapObjName(newObj))
			} else if !utils.IsBool(lazyCache...) {
				DeleteCache(storage, srcDirPath)
			}
		}
	case driver.Rename:
		err = s.Rename(ctx, srcObj, dstName)
		if err == nil && !utils.IsBool(lazyCache...) {
			DeleteCache(storage, srcDirPath)
		}
	default:
		return errors.New("not complemented")
	}
	return errors.WithStack(err)
}

func Copy(ctx context.Context, storage driver.Driver, srcPath, dstDirPath string, lazyCache ...bool) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcPath = utils.FixAndCleanPath(srcPath)
	dstDirPath = utils.FixAndCleanPath(dstDirPath)
	srcObj, err := GetUnwrap(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := GetUnwrap(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get dst dir")
	}

	switch s := storage.(type) {
	case driver.CopyResult:
		var newObj model.Obj
		newObj, err = s.Copy(ctx, srcObj, dstDir)
		if err == nil {
			if newObj != nil {
				addCacheObj(storage, dstDirPath, model.WrapObjName(newObj))
			} else if !utils.IsBool(lazyCache...) {
				DeleteCache(storage, dstDirPath)
			}
		}
	case driver.Copy:
		err = s.Copy(ctx, srcObj, dstDir)
		if err == nil && !utils.IsBool(lazyCache...) {
			DeleteCache(storage, dstDirPath)
		}
	default:
		return errors.New("not complemented")
	}
	return errors.WithStack(err)
}

func Remove(ctx context.Context, storage driver.Driver, path string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	if utils.PathEqual(path, "/") {
		return errors.New("delete root folder is not allowed, please goto the manage page to delete the storage instead")
	}
	path = utils.FixAndCleanPath(path)
	rawObj, err := Get(ctx, storage, path)
	if err != nil {
		if strings.Contains(err.Error(), "object not found") {
			logrus.Debugf("%s have been removed", path)
			return nil
		}
		return errors.WithMessage(err, "failed to get object")
	}
	dirPath := stdpath.Dir(path)

	switch s := storage.(type) {
	case driver.Remove:
		err = s.Remove(ctx, model.UnwrapObj(rawObj))
		if err == nil {
			delCacheObj(storage, dirPath, rawObj)

			// 清除缓存
			if rawObj.IsDir() {
				ClearCache(storage, path)
			}
		}
	default:
		return errors.New("not implemented")
	}
	return errors.WithStack(err)
}

func Other(ctx context.Context, storage driver.Driver, args model.FsOtherArgs) (interface{}, error) {
	obj, err := GetUnwrap(ctx, storage, args.Path)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get obj")
	}
	if o, ok := storage.(driver.Other); ok {
		return o.Other(ctx, model.OtherArgs{
			Obj:    obj,
			Method: args.Method,
			Data:   args.Data,
		})
	} else {
		return nil, errors.New("not implemented")
	}
}

func Put(ctx context.Context, storage driver.Driver, dstDirPath string, file model.FileStreamer, up driver.UpdateProgress, lazyCache ...bool) error {
	close := file.Close
	defer func() {
		if err := close(); err != nil {
			fmt.Errorf("failed to close file streamer, %v", err)
		}
	}()
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	dstDirPath = utils.FixAndCleanPath(dstDirPath)
	dstPath := stdpath.Join(dstDirPath, file.GetName())
	tempName := file.GetName() + ".openlist_to_delete"
	tempPath := stdpath.Join(dstDirPath, tempName)
	fi, err := GetUnwrap(ctx, storage, dstPath)
	if err == nil {
		if fi.GetSize() == 0 {
			err = Remove(ctx, storage, dstPath)
			if err != nil {
				return errors.WithMessagef(err, "while uploading, failed remove existing file which size = 0")
			}
		} else if storage.Config().NoOverwriteUpload {
			// try to rename old obj
			err = Rename(ctx, storage, dstPath, tempName)
			if err != nil {
				return err
			}
		} else {
			file.SetExist(fi)
		}
	}
	err = MakeDir(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to make dir [%s]", dstDirPath)
	}
	parentDir, err := GetUnwrap(ctx, storage, dstDirPath)
	// this should not happen
	if err != nil {
		return errors.WithMessagef(err, "failed to get dir [%s]", dstDirPath)
	}
	if up == nil {
		up = func(p float64) {}
	}

	switch s := storage.(type) {
	case driver.PutResult:
		var newObj model.Obj
		newObj, err = s.Put(ctx, parentDir, file, up)
		if err == nil {
			if newObj != nil {
				addCacheObj(storage, dstDirPath, model.WrapObjName(newObj))
			} else if !utils.IsBool(lazyCache...) {
				DeleteCache(storage, dstDirPath)
			}
		}
	case driver.Put:
		err = s.Put(ctx, parentDir, file, up)
		if err == nil && !utils.IsBool(lazyCache...) {
			DeleteCache(storage, dstDirPath)
		}
	default:
		return fmt.Errorf("NotImplement")
	}
	logrus.Debugf("put file [%s] done", file.GetName())
	if storage.Config().NoOverwriteUpload && fi != nil && fi.GetSize() > 0 {
		if err != nil {
			// 如果上传失败，就恢复旧文件
			err := Rename(ctx, storage, tempPath, file.GetName())
			if err != nil {
				fmt.Errorf("failed recover old obj: %+v", err)
			}
		} else {
			// 如果上传成功，就把旧文件删除
			err := Remove(ctx, storage, tempPath)
			if err != nil {
				return err
			} else {
				// 目前还没实现Link
				// key := Key(storage, stdpath.Join(dstDirPath, file.GetName()))
				// linkCache.Del(key)
			}
		}
	}
	return errors.WithStack(err)
}

var linkCache = cache.NewMemCache(cache.WithShards[*model.Link](16))
var linkG = singleflight.Group[*model.Link]{Remember: true}
var errLinkMFileCache = stderrors.New("ErrLinkMFileCache")

// Link get link, if is an url. should have an expiry time
func Link(ctx context.Context, storage driver.Driver, path string, args model.LinkArgs) (*model.Link, model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != configs.WORK {
		return nil, nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	var (
		file model.Obj
		err  error
	)
	// use cache directly
	dir, name := stdpath.Split(stdpath.Join(storage.GetStorage().MountPath, path))
	if cacheFiles, ok := listCache.Get(strings.TrimSuffix(dir, "/")); ok {
		for _, f := range cacheFiles {
			if f.GetName() == name {
				file = model.UnwrapObj(f)
				break
			}
		}
	} else {
		if g, ok := storage.(driver.GetObjInfo); ok {
			file, err = g.GetObjInfo(ctx, path)
		} else {
			file, err = GetUnwrap(ctx, storage, path)
		}
	}
	if file == nil {
		if err != nil {
			return nil, nil, errors.WithMessage(err, "failed to get file")
		}
		return nil, nil, errors.WithStack(errors.New("ObjectNotFound"))
	}
	if file.IsDir() {
		return nil, nil, errors.WithStack(errors.New("NotFile"))
	}

	key := stdpath.Join(Key(storage, path), args.Type)
	if link, ok := linkCache.Get(key); ok {
		return link, file, nil
	}

	var forget any
	var linkM *model.Link
	fn := func() (*model.Link, error) {
		link, err := storage.Link(ctx, file, args)
		if err != nil {
			return nil, errors.Wrapf(err, "failed get link")
		}
		if link.MFile != nil && forget != nil {
			linkM = link
			return nil, errLinkMFileCache
		}
		if link.Expiration != nil {
			linkCache.Set(key, link, cache.WithEx[*model.Link](*link.Expiration))
		}
		link.AddIfCloser(forget)
		return link, nil
	}

	if storage.Config().OnlyLinkMFile {
		link, err := fn()
		if err != nil {
			return nil, nil, err
		}
		return link, file, err
	}

	forget = utils.CloseFunc(func() error {
		if forget != nil {
			forget = nil
			linkG.Forget(key)
		}
		return nil
	})
	link, err, _ := linkG.Do(key, fn)
	if err == nil && !link.AcquireReference() {
		link, err, _ = linkG.Do(key, fn)
		if err == nil {
			link.AcquireReference()
		}
	}

	if err == errLinkMFileCache {
		if linkM != nil {
			return linkM, file, nil
		}
		forget = nil
		link, err = fn()
	}

	if err != nil {
		return nil, nil, err
	}
	return link, file, nil
}
