package fs

import (
	"HelaList/internal/model"
	"context"
	"fmt"
	"log"
)

// 本目录用于实现fileSystem即文件系统的功能，这也是WebDAV不可缺少的重要组成部分。

type ListArgs struct {
	Refresh bool // 在与远程存储交互时，用缓存来保存上次获取的文件列表
	NoLog   bool // 是否禁止记录错误日志
}

func List(ctx context.Context, path string, args *ListArgs) ([]model.Obj, error) {
	res, err := list(ctx, path, args)
	if err != nil {
		if !args.NoLog {
			log.Println("failed list %s: %+v", path, err)
		}
		return nil, err
	}
	return res, nil
}

func Get(ctx context.Context, path string) (model.Obj, error) {
	res, err := get(ctx, path)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func MakeDir(ctx context.Context, path string, lazyCache ...bool) error {
	err := makeDir(ctx, path, lazyCache...)
	if err != nil {
		fmt.Errorf("failed make dir %s: %+v", path, err)
	}
	return err
}

// func Move(ctx context.Context, srcPath, dstDirPath string, lazyCache ...bool) (task.TaskExtensionInfo, error) {
// 	req, err := transfer(ctx, move, srcPath, dstDirPath, lazyCache...)
// 	if err != nil {
// 		fmt.Errorf("failed move %s to %s: %+v", srcPath, dstDirPath, err)
// 	}
// 	return req, err
// }

// func Copy(ctx context.Context, srcObjPath, dstDirPath string, lazyCache ...bool) (task.TaskExtensionInfo, error) {
// 	res, err := transfer(ctx, copy, srcObjPath, dstDirPath, lazyCache...)
// 	if err != nil {
// 		fmt.Errorf("failed copy %s to %s: %+v", srcObjPath, dstDirPath, err)
// 	}
// 	return res, err
// }

func Rename(ctx context.Context, srcPath, dstName string, lazyCache ...bool) error {
	err := rename(ctx, srcPath, dstName, lazyCache...)
	if err != nil {
		fmt.Errorf("failed rename %s to %s: %+v", srcPath, dstName, err)
	}
	return err
}

func Remove(ctx context.Context, path string) error {
	err := remove(ctx, path)
	if err != nil {
		fmt.Errorf("failed remove %s: %+v", path, err)
	}
	return err
}
