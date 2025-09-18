package driver

import (
	"HelaList/internal/model"
	"context"
)

/*
接口层本身并不难实现，但问题更多地出现在，如何配置一个人人看得懂的接口，
我们看到隔壁的OpenList能够兼容几乎所有的网盘项目，也正因如此，他的接口设计才
极具抽象性，可读性极差。
*/

type Config struct {
	Name      string `json:"name"`
	LocalSort bool   `json:"local_sort"`
	// if the driver returns Link with MFile, this should be set to true
	OnlyLinkMFile bool `json:"only_local"`
	// if the driver can only be proxy
	OnlyProxy bool `json:"only_proxy"`
	NoCache   bool `json:"no_cache"`
	NoUpload  bool `json:"no_upload"`
	// if need get message from user, such as validate code
	NeedMs      bool   `json:"need_ms"`
	DefaultRoot string `json:"default_root"`
	CheckStatus bool   `json:"-"`
	//info,success,warning,danger
	Alert string `json:"alert"`
	// whether to support overwrite upload
	NoOverwriteUpload bool `json:"-"`
}

// 进行文件代理
func (c Config) MustProxy() bool {
	return c.OnlyProxy || c.OnlyLinkMFile
}

// 对一个网盘的接口抽象
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
	Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error)
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

type Put interface {
	Put(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up UpdateProgress) error
}

type Reference interface {
	InitReference(storage Driver) error
}

// 以下为带结果返回，唯一区别是返回对应Obj

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

type PutResult interface {
	Put(ctx context.Context, dstDir model.Obj, file model.FileStreamer, up UpdateProgress) (model.Obj, error)
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
