package model

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/OpenListTeam/OpenList/v4/pkg/http_range"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
)

type FileCloser struct {
	File
	io.Closer
}

type FileRangeReader struct {
	RangeReaderIF
}

// 用于在分享文件时生成URL，文件分享功能待制作
type Link struct {
	URL         string        `json:"url"`    // most common way
	Header      http.Header   `json:"header"` // needed header (for url)
	RangeReader RangeReaderIF `json:"-"`      // recommended way if can't use URL
	MFile       File          `json:"-"`      // best for local,smb... file system, which exposes MFile

	Expiration *time.Duration // local cache expire Duration

	//for accelerating request, use multi-thread downloading
	Concurrency   int   `json:"concurrency"`
	PartSize      int   `json:"part_size"`
	ContentLength int64 `json:"-"` // 转码视频、缩略图

	utils.SyncClosers `json:"-"`
}

// File表示一个支持顺序访问和随机访问的文件，本身为接口，需要依靠底层实现
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// FileStreamer是FileStream的接口
/*
不过你不写这个接口其实也可以，因为FileStreamer的实现方案只有FileStream一种
*/
type FileStreamer interface {
	io.Reader
	utils.ClosersIF
	Obj
	GetMimetype() string
	NeedStore() bool
	IsForceStreamUpload() bool
	GetExist() Obj
	SetExist(Obj)
	RangeRead(http_range.Range) (io.Reader, error)
	// if the Stream is not a File and is not cached, returns nil.
	GetFile() File

	// 与后续写文件相关
	// for a non-seekable Stream, if Read is called, this function won't work.
	// caches the full Stream and writes it to writer (if provided, even if the stream is already cached).
	CacheFullAndWriter(up *UpdateProgress, writer io.Writer) (File, error)
	SetTmpFile(file File)
}

func (f *FileCloser) Close() error {
	var errs []error
	if clr, ok := f.File.(io.Closer); ok {
		errs = append(errs, clr.Close())
	}
	if f.Closer != nil {
		errs = append(errs, f.Closer.Close())
	}
	return errors.Join(errs...)
}

type UpdateProgress func(percentage float64)

func UpdateProgressWithRange(inner UpdateProgress, start, end float64) UpdateProgress {
	return func(p float64) {
		if p < 0 {
			p = 0
		}
		if p > 100 {
			p = 100
		}
		scaled := start + (end-start)*(p/100.0)
		inner(scaled)
	}
}
