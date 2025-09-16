package model

import (
	"time"

	"github.com/google/uuid"
)

type Object struct {
	Id           uuid.UUID
	Path         string
	Name         string
	Size         int64
	ModifiedTime time.Time
	CreatedTime  time.Time // 文件创建时间
	IsFolder     bool
	// HashInfo // 哈希检验和先不做
}

func (o *Object) GetName() string {
	return o.Name
}

func (o *Object) GetSize() int64 {
	return o.Size
}

func (o *Object) GetModifiedTime() time.Time {
	return o.ModifiedTime
}

func (o *Object) GetCreatedTime() time.Time {
	if o.CreatedTime.IsZero() {
		return o.GetModifiedTime()
	}
	return o.CreatedTime
}

func (o *Object) IsDir() bool {
	return o.IsFolder
}

func (o *Object) GetId() uuid.UUID {
	return o.Id
}

func (o *Object) GetPath() string {
	return o.Path
}

func (o *Object) SetPath(path string) {
	o.Path = path
}

/*
func (o *Object) GetHash()
*/
