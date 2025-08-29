package fs

import (
	"HelaList/internal/model"
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
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

func Move(ctx context.Context, srcPath, dstPath string, lazyCache ...bool) error {
	err := move(ctx, srcPath, dstPath, lazyCache...)
	if err != nil {
		// 遵循该文件中其他函数的错误处理风格
		fmt.Errorf("failed move %s to %s: %+v", srcPath, dstPath, err)
	}
	return err
}

func Copy(ctx context.Context, srcPath, dstPath string, lazyCache ...bool) error {
	err := copy(ctx, srcPath, dstPath, lazyCache...)
	if err != nil {
		// 遵循该文件中其他函数的错误处理风格
		fmt.Errorf("failed copy %s to %s: %+v", srcPath, dstPath, err)
	}
	return err
}

// PutDirectly 将文件直接上传并等待完成。
func PutDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer, lazyCache ...bool) error {
	err := putDirectly(ctx, dstDirPath, file, lazyCache...)
	if err != nil {
		// 使用 logrus 记录错误日志，与您提供的 fs/fs.go 文件保持一致
		logrus.Errorf("failed put %s: %+v", dstDirPath, err)
	}
	return err
}
