package fs

import (
	"HelaList/internal/model"
	"context"
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
