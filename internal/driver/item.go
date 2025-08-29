package driver

import "github.com/google/uuid"

// item负责实现数据库的配置字段

// 用于JSON
type Additional interface{}

type Item struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Options  string `json:"options"`
	Required bool   `json:"required"`
	Help     string `json:"help"`
}

type Info struct {
	Common     []Item `json:"common"`
	Additional []Item `json:"additional"`
	Config     Config `json:"config"`
}

type RootPath struct {
	RootFolderPath string `json:"root_folder_path"`
}

type RootID struct {
	RootFolderID string `json:"root_folder_id"`
}

// 获取根目录
type IRootPath interface {
	GetRootPath() string
}

// 获取根目录的Id
type IRootId interface {
	GetRootId() uuid.UUID
}

func (r RootPath) GetRootPath() string {
	return r.RootFolderPath
}

func (r *RootPath) SetRootPath(path string) {
	r.RootFolderPath = path
}

func (r RootID) GetRootId() string {
	return r.RootFolderID
}

// 对应服务器端网盘的挂载目录
/*
切记不能挂载到根目录，否则可能会出现根目录权限问题
项目早期测试以坚果云作为服务器端，部署在了/HelaList目录下而非/根目录，
就是因为权限问题，绑定根目录会导致坚果云一直返回403
*/
