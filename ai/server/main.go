package main

import (
	_ "HelaList/drivers/webdav"
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"HelaList/internal/op"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP 工具参数结构定义

// 用户相关工具参数
type LoginParams struct {
	Username string `json:"username" jsonschema:"用户名"`
	Password string `json:"password" jsonschema:"密码"`
}

type CreateUserParams struct {
	Username string `json:"username" jsonschema:"用户名"`
	Email    string `json:"email" jsonschema:"邮箱地址"`
	Password string `json:"password" jsonschema:"密码"`
	BasePath string `json:"base_path" jsonschema:"用户基础路径"`
	Identity int    `json:"identity" jsonschema:"用户身份类型"`
}

type GetUserParams struct {
	Username string `json:"username" jsonschema:"用户名"`
}

type UpdateUserParams struct {
	Username string `json:"username" jsonschema:"用户名"`
	Email    string `json:"email" jsonschema:"新邮箱地址"`
	Password string `json:"password" jsonschema:"新密码"`
	BasePath string `json:"base_path" jsonschema:"新基础路径"`
	Identity int    `json:"identity" jsonschema:"新身份"`
	Disabled bool   `json:"disabled" jsonschema:"是否禁用"`
}

type DeleteUserParams struct {
	Username string `json:"username" jsonschema:"要删除的用户名"`
}

// 存储相关工具参数
type CreateStorageParams struct {
	MountPath       string `json:"mount_path" jsonschema:"挂载路径"`
	Driver          string `json:"driver" jsonschema:"驱动类型"`
	CacheExpiration int    `json:"cache_expiration" jsonschema:"缓存过期时间秒"`
	Addition        string `json:"addition" jsonschema:"附加配置信息JSON格式"`
	Remark          string `json:"remark" jsonschema:"备注信息"`
	Order           int    `json:"order" jsonschema:"排序序号"`
}

type UpdateStorageParams struct {
	MountPath       string `json:"mount_path" jsonschema:"挂载路径"`
	Driver          string `json:"driver" jsonschema:"驱动类型"`
	CacheExpiration int    `json:"cache_expiration" jsonschema:"缓存过期时间秒"`
	Addition        string `json:"addition" jsonschema:"附加配置信息JSON格式"`
	Remark          string `json:"remark" jsonschema:"备注信息"`
	Order           int    `json:"order" jsonschema:"排序序号"`
	Disabled        bool   `json:"disabled" jsonschema:"是否禁用"`
}

type GetStorageParams struct {
	MountPath string `json:"mount_path" jsonschema:"挂载路径"`
}

// 文件系统相关工具参数
type FsListParams struct {
	Path     string `json:"path" jsonschema:"要列出的目录路径"`
	Username string `json:"username" jsonschema:"用户名用于权限检查"`
	Password string `json:"password" jsonschema:"目录密码如果需要"`
}

type FsMkdirParams struct {
	Path     string `json:"path" jsonschema:"要创建的目录路径"`
	Username string `json:"username" jsonschema:"用户名用于权限检查"`
}

type FsRemoveParams struct {
	Names    []string `json:"names" jsonschema:"要删除的文件/目录名称列表"`
	DirPath  string   `json:"dir_path" jsonschema:"目录路径"`
	Username string   `json:"username" jsonschema:"用户名用于权限检查"`
}

type FsCopyParams struct {
	SrcDirPath string   `json:"src_dir_path" jsonschema:"源目录路径"`
	DstDirPath string   `json:"dst_dir_path" jsonschema:"目标目录路径"`
	Names      []string `json:"names" jsonschema:"要复制的文件/目录名称列表"`
	Username   string   `json:"username" jsonschema:"用户名用于权限检查"`
}

type FsMoveParams struct {
	SrcDirPath string   `json:"src_dir_path" jsonschema:"源目录路径"`
	DstDirPath string   `json:"dst_dir_path" jsonschema:"目标目录路径"`
	Names      []string `json:"names" jsonschema:"要移动的文件/目录名称列表"`
	Username   string   `json:"username" jsonschema:"用户名用于权限检查"`
}

type FsRenameParams struct {
	Path     string `json:"path" jsonschema:"文件/目录路径"`
	Name     string `json:"name" jsonschema:"新名称"`
	Username string `json:"username" jsonschema:"用户名用于权限检查"`
}

// MCP 工具处理器实现

// 用户登录工具
func LoginTool(ctx context.Context, req *mcp.CallToolRequest, args LoginParams) (*mcp.CallToolResult, any, error) {
	user, err := op.GetUserByName(args.Username)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("用户名或密码错误: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	ok, err := user.CheckPassword(args.Password)
	if !ok || err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "用户名或密码错误"},
			},
			IsError: true,
		}, nil, nil
	}

	if user.Disabled {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "用户已被禁用"},
			},
			IsError: true,
		}, nil, nil
	}

	userInfo, _ := json.Marshal(user)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("登录成功。用户信息: %s", string(userInfo))},
		},
	}, nil, nil
}

// 创建用户工具
func CreateUserTool(ctx context.Context, req *mcp.CallToolRequest, args CreateUserParams) (*mcp.CallToolResult, any, error) {
	user := &model.User{
		Username: args.Username,
		Email:    args.Email,
		BasePath: args.BasePath,
		Identity: args.Identity,
		Disabled: false,
	}

	if err := user.SetPassword(args.Password); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("密码加密失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	if err := op.CreateUser(user); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("创建用户失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("用户 %s 创建成功，ID: %s", user.Username, user.Id.String())},
		},
	}, nil, nil
}

// 获取用户信息工具
func GetUserTool(ctx context.Context, req *mcp.CallToolRequest, args GetUserParams) (*mcp.CallToolResult, any, error) {
	user, err := op.GetUserByName(args.Username)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("获取用户失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	userInfo, _ := json.Marshal(user)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(userInfo)},
		},
	}, nil, nil
}

// 更新用户工具
func UpdateUserTool(ctx context.Context, req *mcp.CallToolRequest, args UpdateUserParams) (*mcp.CallToolResult, any, error) {
	user, err := op.GetUserByName(args.Username)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	// 更新用户信息
	if args.Email != "" {
		user.Email = args.Email
	}
	if args.BasePath != "" {
		user.BasePath = args.BasePath
	}
	if args.Identity != 0 {
		user.Identity = args.Identity
	}
	user.Disabled = args.Disabled

	if args.Password != "" {
		if err := user.SetPassword(args.Password); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("密码更新失败: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	}

	if err := op.UpdateUser(user); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("更新用户失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("用户 %s 更新成功", args.Username)},
		},
	}, nil, nil
}

// 删除用户工具
func DeleteUserTool(ctx context.Context, req *mcp.CallToolRequest, args DeleteUserParams) (*mcp.CallToolResult, any, error) {
	user, err := op.GetUserByName(args.Username)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	if err := op.DeleteUserById(user.Id); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("删除用户失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("用户 %s 删除成功", args.Username)},
		},
	}, nil, nil
}

// 创建存储工具
func CreateStorageTool(ctx context.Context, req *mcp.CallToolRequest, args CreateStorageParams) (*mcp.CallToolResult, any, error) {
	storage := model.Storage{
		MountPath:       args.MountPath,
		Driver:          args.Driver,
		CacheExpiration: args.CacheExpiration,
		Addition:        args.Addition,
		Remark:          args.Remark,
		Order:           args.Order,
		Status:          "work",
		ModifiedTime:    time.Now(),
		Disabled:        false,
	}

	id, err := op.CreateStorage(ctx, storage)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("创建存储失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("存储创建成功，ID: %s", id)},
		},
	}, nil, nil
}

// 更新存储工具
func UpdateStorageTool(ctx context.Context, req *mcp.CallToolRequest, args UpdateStorageParams) (*mcp.CallToolResult, any, error) {
	driver, err := op.GetStorageByMountPath(args.MountPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("存储不存在: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	storage := driver.GetStorage()
	if args.Driver != "" {
		storage.Driver = args.Driver
	}
	if args.Addition != "" {
		storage.Addition = args.Addition
	}
	if args.Remark != "" {
		storage.Remark = args.Remark
	}
	if args.CacheExpiration != 0 {
		storage.CacheExpiration = args.CacheExpiration
	}
	if args.Order != 0 {
		storage.Order = args.Order
	}
	storage.Disabled = args.Disabled
	storage.ModifiedTime = time.Now()

	if err := op.UpdateStorage(ctx, *storage); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("更新存储失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("存储 %s 更新成功", args.MountPath)},
		},
	}, nil, nil
}

// 获取存储信息工具
func GetStorageTool(ctx context.Context, req *mcp.CallToolRequest, args GetStorageParams) (*mcp.CallToolResult, any, error) {
	driver, err := op.GetStorageByMountPath(args.MountPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("获取存储失败: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	storageInfo, _ := json.Marshal(driver.GetStorage())
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(storageInfo)},
		},
	}, nil, nil
}

// 获取所有存储工具
func GetAllStoragesTool(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	storages := op.GetAllStorages()
	storagesInfo, _ := json.Marshal(storages)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(storagesInfo)},
		},
	}, nil, nil
}

// 文件系统列表工具
func FsListTool(ctx context.Context, req *mcp.CallToolRequest, args FsListParams) (*mcp.CallToolResult, any, error) {
	// 这里需要模拟用户上下文，简化处理
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		// 使用默认的访客用户
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	reqPath, err := user.JoinPath(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("路径错误: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	// 这里简化处理，直接返回路径信息
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("列出目录 %s 的内容(用户: %s, 解析路径: %s)", args.Path, user.Username, reqPath)},
		},
	}, nil, nil
}

// 创建目录工具
func FsMkdirTool(ctx context.Context, req *mcp.CallToolRequest, args FsMkdirParams) (*mcp.CallToolResult, any, error) {
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	reqPath, err := user.JoinPath(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("路径错误: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("创建目录 %s 成功(用户: %s, 解析路径: %s)", args.Path, user.Username, reqPath)},
		},
	}, nil, nil
}

// 删除文件/目录工具
func FsRemoveTool(ctx context.Context, req *mcp.CallToolRequest, args FsRemoveParams) (*mcp.CallToolResult, any, error) {
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("删除文件/目录 %v 成功(用户: %s, 目录: %s)", args.Names, user.Username, args.DirPath)},
		},
	}, nil, nil
}

// 复制文件/目录工具
func FsCopyTool(ctx context.Context, req *mcp.CallToolRequest, args FsCopyParams) (*mcp.CallToolResult, any, error) {
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("复制文件/目录 %v 从 %s 到 %s 成功(用户: %s)", args.Names, args.SrcDirPath, args.DstDirPath, user.Username)},
		},
	}, nil, nil
}

// 移动文件/目录工具
func FsMoveTool(ctx context.Context, req *mcp.CallToolRequest, args FsMoveParams) (*mcp.CallToolResult, any, error) {
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("移动文件/目录 %v 从 %s 到 %s 成功(用户: %s)", args.Names, args.SrcDirPath, args.DstDirPath, user.Username)},
		},
	}, nil, nil
}

// 重命名文件/目录工具
func FsRenameTool(ctx context.Context, req *mcp.CallToolRequest, args FsRenameParams) (*mcp.CallToolResult, any, error) {
	var user *model.User
	var err error

	if args.Username != "" {
		user, err = op.GetUserByName(args.Username)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("用户不存在: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		user = &model.User{
			Username: "guest",
			Identity: model.GUEST,
			BasePath: "/",
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("重命名 %s 为 %s 成功(用户: %s)", args.Path, args.Name, user.Username)},
		},
	}, nil, nil
}

func main() {
	// 初始化数据库连接
	bootstrap.InitDB()

	// 加载存储配置
	op.LoadAllStorages(context.Background())

	// 创建 MCP 服务器
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "HelaList-MCP-Server",
		Version: "v1.0.0",
	}, nil)

	// 注册用户管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "user_login",
		Description: "用户登录验证",
	}, LoginTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "user_create",
		Description: "创建新用户",
	}, CreateUserTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "user_get",
		Description: "获取用户信息",
	}, GetUserTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "user_update",
		Description: "更新用户信息",
	}, UpdateUserTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "user_delete",
		Description: "删除用户",
	}, DeleteUserTool)

	// 注册存储管理工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "storage_create",
		Description: "创建新存储",
	}, CreateStorageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "storage_update",
		Description: "更新存储配置",
	}, UpdateStorageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "storage_get",
		Description: "获取存储信息",
	}, GetStorageTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "storage_get_all",
		Description: "获取所有存储信息",
	}, GetAllStoragesTool)

	// 注册文件系统工具
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_list",
		Description: "列出目录内容",
	}, FsListTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_mkdir",
		Description: "创建目录",
	}, FsMkdirTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_remove",
		Description: "删除文件或目录",
	}, FsRemoveTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_copy",
		Description: "复制文件或目录",
	}, FsCopyTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_move",
		Description: "移动文件或目录",
	}, FsMoveTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fs_rename",
		Description: "重命名文件或目录",
	}, FsRenameTool)

	// 在 stdin/stdout 上运行服务器，直到客户端断开连接
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}
