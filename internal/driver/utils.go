package driver

// 驱动接口层部分功能函数

import (
	"HelaList/internal/model"
	"HelaList/internal/stream"
	"context"
	"io"
)

// 进度条相关，在涉及文件上传/下载时发挥作用

type UpdateProgress = model.UpdateProgress

type Progress struct {
	Total int64
	Done  int64
	up    UpdateProgress
}

func (p *Progress) Write(b []byte) (n int, err error) {
	n = len(b)
	p.Done += int64(n)
	p.up(float64(p.Done) / float64(p.Total) * 100)
	return
}

func NewProgress(total int64, up UpdateProgress) *Progress {
	return &Progress{
		Total: total,
		up:    up,
	}
}

// 文件流相关

type RateLimitReader = stream.RateLimitReader

type RateLimitWriter = stream.RateLimitWriter

type RateLimitFile = stream.RateLimitFile

type ReaderWithCtx = stream.ReaderWithCtx

type ReaderUpdatingProgress = stream.ReaderUpdatingProgress

type SimpleReaderWithSize = stream.SimpleReaderWithSize

func NewLimitedUploadStream(ctx context.Context, r io.Reader) *RateLimitReader {
	return &RateLimitReader{
		Reader:  r,
		Limiter: stream.ServerUploadLimit,
		Ctx:     ctx,
	}
}
