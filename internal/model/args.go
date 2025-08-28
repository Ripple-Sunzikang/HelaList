package model

import (
	"context"
	"io"

	"github.com/OpenListTeam/OpenList/v4/pkg/http_range"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
)

// 存放一堆参数用的文件

type ListArgs struct {
	ReqPath string
	Refresh bool
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
