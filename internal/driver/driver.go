package driver

import (
	"HelaList/internal/model"
	"context"

	"github.com/google/uuid"
)

type Driver interface {
	Meta   // 网盘元数据
	Reader // 网盘读取操作
}

type Meta interface {
	/*
		GetStorage()这玩意儿是需要实现的吗？
		毕竟model.Storage已经有GetStorage了，不知道的还以为是跨命名空间接口实现呢
	*/
	GetStorage() *model.Storage
	SetStorage(model.Storage)
	//
	GetAddition() Additional
}

// 用于JSON
type Additional interface{}

// 用于读取路径下的所有文件
type Reader interface {
	List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error)
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

type IRootPath interface {
	GetRootPath() string
}

type IRootId interface {
	GetRootId() uuid.UUID
}

type RootPath struct {
	RootFolderPath string `json:"root_folder_path"`
}

type RootID struct {
	RootFolderID uuid.UUID `json:"root_folder_id"`
}

func (r RootPath) GetRootPath() string {
	return r.RootFolderPath
}

func (r *RootPath) SetRootPath(path string) {
	r.RootFolderPath = path
}

func (r RootID) GetRootId() uuid.UUID {
	return r.RootFolderID
}
