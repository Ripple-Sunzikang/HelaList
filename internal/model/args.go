package model

import (
	"context"
	"io"
	"net/http"

	"github.com/OpenListTeam/OpenList/v4/pkg/http_range"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
)

// 我有点后悔写args.go了，因为这里就是个垃圾场。

type ListArgs struct {
	ReqPath string
	Refresh bool
}

type LinkArgs struct {
	IP       string
	Header   http.Header
	Type     string
	Redirect bool
}

type FsOtherArgs struct {
	Path   string      `json:"path" form:"path"`
	Method string      `json:"method" form:"method"`
	Data   interface{} `json:"data" form:"data"`
}

type OtherArgs struct {
	Obj    Obj
	Method string
	Data   interface{}
}

// 指定范围读取文件相关接口，用于后续实现随机访问文件

type RangeReaderIF interface {
	RangeRead(ctx context.Context, httpRange http_range.Range) (io.ReadCloser, error)
}

type RangeReadCloserIF interface {
	RangeReaderIF
	utils.ClosersIF
}

var _ RangeReadCloserIF = (*RangeReadCloser)(nil)

type RangeReadCloser struct {
	RangeReader RangeReaderIF
	utils.Closers
}

func (r *RangeReadCloser) RangeRead(ctx context.Context, httpRange http_range.Range) (io.ReadCloser, error) {
	rc, err := r.RangeReader.RangeRead(ctx, httpRange)
	r.Add(rc)
	return rc, err
}
