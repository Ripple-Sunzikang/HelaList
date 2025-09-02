package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// 客户端配置
type ClientConfig struct {
	ServerCommand string   `json:"server_command"`
	ServerArgs    []string `json:"server_args"`
	Timeout       int      `json:"timeout_seconds"`
}

// 加载配置文件
func loadConfig() (*ClientConfig, error) {
	configFile := "client_config.json"

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := &ClientConfig{
			ServerCommand: "go",
			ServerArgs:    []string{"run", "../server/main.go"},
			Timeout:       30,
		}

		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}

		log.Printf("已创建默认配置文件: %s", configFile)
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// 测试工具调用的辅助函数
func callTool(session *mcp.ClientSession, toolName string, args map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	}

	result, err := session.CallTool(ctx, params)
	if err != nil {
		return fmt.Errorf("调用工具 %s 失败: %v", toolName, err)
	}

	if result.IsError {
		log.Printf("工具 %s 执行错误", toolName)
	} else {
		log.Printf("工具 %s 执行成功", toolName)
	}

	// 打印结果内容
	for i, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			log.Printf("结果 %d: %s", i+1, textContent.Text)
		}
	}

	return nil
}

// 列出可用工具
func listTools(session *mcp.ClientSession) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取工具列表失败: %v", err)
	}

	log.Println("可用工具列表:")
	for _, tool := range result.Tools {
		log.Printf("  - %s: %s", tool.Name, tool.Description)
	}

	return nil
}

// 测试用户管理功能
func testUserManagement(session *mcp.ClientSession) {
	log.Println("\n=== 测试用户管理功能 ===")

	// 1. 创建测试用户
	log.Println("1. 创建测试用户...")
	err := callTool(session, "user_create", map[string]any{
		"username":  "testuser",
		"email":     "test@example.com",
		"password":  "testpass123",
		"base_path": "/home/testuser",
		"identity":  1, // 普通用户
	})
	if err != nil {
		log.Printf("创建用户失败: %v", err)
	}

	// 2. 获取用户信息
	log.Println("2. 获取用户信息...")
	err = callTool(session, "user_get", map[string]any{
		"username": "testuser",
	})
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
	}

	// 3. 测试用户登录
	log.Println("3. 测试用户登录...")
	err = callTool(session, "user_login", map[string]any{
		"username": "testuser",
		"password": "testpass123",
	})
	if err != nil {
		log.Printf("用户登录失败: %v", err)
	}

	// 4. 更新用户信息
	log.Println("4. 更新用户信息...")
	err = callTool(session, "user_update", map[string]any{
		"username":  "testuser",
		"email":     "newemail@example.com",
		"base_path": "/home/testuser/new",
	})
	if err != nil {
		log.Printf("更新用户失败: %v", err)
	}
}

// 测试存储管理功能
func testStorageManagement(session *mcp.ClientSession) {
	log.Println("\n=== 测试存储管理功能 ===")

	// 1. 创建测试存储
	log.Println("1. 创建测试存储...")
	err := callTool(session, "storage_create", map[string]any{
		"mount_path":       "/test-storage",
		"driver":           "webdav",
		"cache_expiration": 3600,
		"addition":         `{"endpoint": "http://localhost:8080", "username": "test", "password": "test"}`,
		"remark":           "测试存储",
		"order":            1,
	})
	if err != nil {
		log.Printf("创建存储失败: %v", err)
	}

	// 2. 获取存储信息
	log.Println("2. 获取存储信息...")
	err = callTool(session, "storage_get", map[string]any{
		"mount_path": "/test-storage",
	})
	if err != nil {
		log.Printf("获取存储信息失败: %v", err)
	}

	// 3. 获取所有存储
	log.Println("3. 获取所有存储...")
	err = callTool(session, "storage_get_all", map[string]any{})
	if err != nil {
		log.Printf("获取所有存储失败: %v", err)
	}

	// 4. 更新存储信息
	log.Println("4. 更新存储信息...")
	err = callTool(session, "storage_update", map[string]any{
		"mount_path": "/test-storage",
		"remark":     "更新后的测试存储",
		"order":      2,
	})
	if err != nil {
		log.Printf("更新存储失败: %v", err)
	}
}

// 测试文件系统功能
func testFileSystemOperations(session *mcp.ClientSession) {
	log.Println("\n=== 测试文件系统功能 ===")

	// 1. 列出根目录
	log.Println("1. 列出根目录...")
	err := callTool(session, "fs_list", map[string]any{
		"path":     "/",
		"username": "testuser",
	})
	if err != nil {
		log.Printf("列出目录失败: %v", err)
	}

	// 2. 创建目录
	log.Println("2. 创建测试目录...")
	err = callTool(session, "fs_mkdir", map[string]any{
		"path":     "/test-dir",
		"username": "testuser",
	})
	if err != nil {
		log.Printf("创建目录失败: %v", err)
	}

	// 3. 重命名操作
	log.Println("3. 重命名文件...")
	err = callTool(session, "fs_rename", map[string]any{
		"path":     "/test-dir",
		"name":     "renamed-test-dir",
		"username": "testuser",
	})
	if err != nil {
		log.Printf("重命名失败: %v", err)
	}

	// 4. 复制操作
	log.Println("4. 复制文件...")
	err = callTool(session, "fs_copy", map[string]any{
		"src_dir_path": "/",
		"dst_dir_path": "/backup",
		"names":        []string{"renamed-test-dir"},
		"username":     "testuser",
	})
	if err != nil {
		log.Printf("复制失败: %v", err)
	}

	// 5. 移动操作
	log.Println("5. 移动文件...")
	err = callTool(session, "fs_move", map[string]any{
		"src_dir_path": "/backup",
		"dst_dir_path": "/archive",
		"names":        []string{"renamed-test-dir"},
		"username":     "testuser",
	})
	if err != nil {
		log.Printf("移动失败: %v", err)
	}

	// 6. 删除操作
	log.Println("6. 删除文件...")
	err = callTool(session, "fs_remove", map[string]any{
		"names":    []string{"renamed-test-dir"},
		"dir_path": "/archive",
		"username": "testuser",
	})
	if err != nil {
		log.Printf("删除失败: %v", err)
	}
}

// 清理测试数据
func cleanupTestData(session *mcp.ClientSession) {
	log.Println("\n=== 清理测试数据 ===")

	// 删除测试用户
	log.Println("删除测试用户...")
	err := callTool(session, "user_delete", map[string]any{
		"username": "testuser",
	})
	if err != nil {
		log.Printf("删除测试用户失败: %v", err)
	}
}

// 交互式模式
func interactiveMode(session *mcp.ClientSession) {
	log.Println("\n=== 进入交互式模式 ===")
	log.Println("输入 'help' 查看可用命令，输入 'quit' 退出")

	for {
		fmt.Print("MCP> ")

		var input string
		fmt.Scanln(&input)

		switch input {
		case "quit", "exit":
			log.Println("退出交互式模式")
			return
		case "help":
			log.Println("可用命令:")
			log.Println("  help          - 显示此帮助信息")
			log.Println("  tools         - 列出所有可用工具")
			log.Println("  test-user     - 测试用户管理功能")
			log.Println("  test-storage  - 测试存储管理功能")
			log.Println("  test-fs       - 测试文件系统功能")
			log.Println("  cleanup       - 清理测试数据")
			log.Println("  quit/exit     - 退出")
		case "tools":
			if err := listTools(session); err != nil {
				log.Printf("错误: %v", err)
			}
		case "test-user":
			testUserManagement(session)
		case "test-storage":
			testStorageManagement(session)
		case "test-fs":
			testFileSystemOperations(session)
		case "cleanup":
			cleanupTestData(session)
		default:
			log.Printf("未知命令: %s，输入 'help' 查看可用命令", input)
		}
	}
}

func main() {
	log.Println("HelaList MCP 客户端启动中...")

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建客户端
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "HelaList-MCP-Client",
		Version: "v1.0.0",
	}, &mcp.ClientOptions{
		KeepAlive: 30 * time.Second, // 30秒保活
	})

	// 创建命令传输（连接到服务器）
	cmdArgs := append([]string{config.ServerCommand}, config.ServerArgs...)
	transport := &mcp.CommandTransport{
		Command: exec.Command(cmdArgs[0], cmdArgs[1:]...),
	}

	log.Printf("连接到服务器: %s %v", config.ServerCommand, config.ServerArgs)

	// 连接到服务器
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("连接服务器失败: %v", err)
	}
	defer session.Close()

	log.Println("成功连接到 MCP 服务器!")

	// 列出可用工具
	if err := listTools(session); err != nil {
		log.Printf("获取工具列表失败: %v", err)
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "test":
			// 运行所有测试
			log.Println("\n开始运行完整测试套件...")
			testUserManagement(session)
			testStorageManagement(session)
			testFileSystemOperations(session)
			cleanupTestData(session)
			log.Println("\n测试完成!")
		case "test-user":
			testUserManagement(session)
		case "test-storage":
			testStorageManagement(session)
		case "test-fs":
			testFileSystemOperations(session)
		case "interactive":
			interactiveMode(session)
		default:
			log.Printf("未知参数: %s", os.Args[1])
			log.Println("可用参数: test, test-user, test-storage, test-fs, interactive")
		}
	} else {
		// 默认进入交互式模式
		interactiveMode(session)
	}

	log.Println("客户端结束运行")
}
