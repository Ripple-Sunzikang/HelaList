package model

import "time"

/*
严格来说，Obj接口才是文件最抽象的形式。其次才是Object。
*/
type Obj interface {
	GetSize() int64
	GetName() string
	GetModifiedTime() time.Time
	GetCreatedTime() time.Time
	IsDir() bool
	// GetHash() // 哈希还没实现，所以暂时不做
}
