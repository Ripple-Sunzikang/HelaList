package model

import (
	"time"
)

type Object struct {
	Id           string
	Path         string
	Name         string
	Size         int64
	ModifiedTime time.Time
	CreateTime   time.Time // 文件创建时间
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

func (o *Object) GetCreateTime() time.Time {
	if o.CreateTime.IsZero() {
		return o.GetModifiedTime()
	}
	return o.CreateTime
}

func (o *Object) IsDir() bool {
	return o.IsFolder
}

func (o *Object) GetId() string {
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
