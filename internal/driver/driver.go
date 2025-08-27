package driver

import (
	"HelaList/internal/model"
	"context"
)

type Config struct {
	Name      string `json:"name"`
	LocalSort bool   `json:"local_sort"`
	// if the driver returns Link with MFile, this should be set to true
	OnlyLinkMFile bool `json:"only_local"`
	OnlyProxy     bool `json:"only_proxy"`
	NoCache       bool `json:"no_cache"`
	NoUpload      bool `json:"no_upload"`
	// if need get message from user, such as validate code
	NeedMs      bool   `json:"need_ms"`
	DefaultRoot string `json:"default_root"`
	CheckStatus bool   `json:"-"`
	//info,success,warning,danger
	Alert string `json:"alert"`
	// whether to support overwrite upload
	NoOverwriteUpload bool `json:"-"`
	ProxyRangeOption  bool `json:"-"`
	// if the driver returns Link without URL, this should be set to true
	NoLinkURL bool `json:"-"`
}

type Driver interface {
	Meta   // 网盘元数据
	Reader // 网盘读取操作
}

type Meta interface {
	/*
		GetStorage()这玩意儿是需要实现的吗？
		毕竟model.Storage已经有GetStorage了，不知道的还以为是跨命名空间接口实现呢
	*/
	Config() Config
	GetStorage() *model.Storage
	SetStorage(model.Storage)
	//
	GetAddition() Additional
	Init(ctx context.Context) error
	Drop(ctx context.Context) error
}

// 用于读取路径下的所有文件
type Reader interface {
	List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error)
	// Link(ctx context.Context, file model.Obj, args model.ListArgs) (*model.Link, error)
}

// 获取根目录
type GetRooter interface {
	GetRoot(ctx context.Context) (model.Obj, error)
}

// 通过路径查找文件
type Getter interface {
	Get(ctx context.Context, path string) (model.Obj, error)
}

// 通过文件路径，检索文件信息
type GetObjInfo interface {
	GetObjInfo(ctx context.Context, path string) (model.Obj, error)
}

// 写操作相关接口(好多)
/*
注意，因为Obj包含了文件/文件夹两种情况，
所以，为了区分二者，在函数参数里，有时命名会刻意区分Obj和Dir
*/
type Mkdir interface {
	MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error
}

type Move interface {
	Move(ctx context.Context, sourceObj model.Obj, destiDir model.Obj) error
}

type Rename interface {
	Rename(ctx context.Context, obj model.Obj, newName string) error
}

type Copy interface {
	Copy(ctx context.Context, sourceObj model.Obj, destiDir model.Obj) error
}

type Remove interface {
	Remove(ctx context.Context, obj model.Obj) error
}

/*
// Put用于上传文件，而上传/下载文件这种事往往复杂，还需要做文件流
type Put interface {
	Put(ctx context.Context, destiDIr model.Obj, )
}
*/

type Reference interface {
	InitReference(storage Driver) error
}

//type WriteResult interface {
//	MkdirResult
//	MoveResult
//	RenameResult
//	CopyResult
//	PutResult
//	Remove
//}

type MkdirResult interface {
	MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error)
}

type MoveResult interface {
	Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error)
}

type RenameResult interface {
	Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error)
}

type CopyResult interface {
	Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error)
}

type PutURLResult interface {
	// PutURL directly put a URL into the storage
	// Applicable to index-based drivers like URL-Tree or drivers that support uploading files as URLs
	// Called when using SimpleHttp for offline downloading, skipping creating a download task
	PutURL(ctx context.Context, dstDir model.Obj, name, url string) (model.Obj, error)
}

type Other interface {
	Other(ctx context.Context, args model.OtherArgs) (interface{}, error)
}
