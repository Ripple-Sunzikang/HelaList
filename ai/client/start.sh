#!/bin/bash

# HelaList MCP 启动脚本

echo "HelaList MCP 系统启动脚本"
echo "========================"

# 检查是否在正确的目录
if [ ! -f "../../go.mod" ]; then
    echo "错误: 请在 ai/client 目录下运行此脚本"
    exit 1
fi

# 确保依赖已安装
echo "检查依赖..."
cd ../..
go mod tidy

# 返回客户端目录
cd ai/client

echo ""
echo "可用选项:"
echo "1. 运行完整测试套件"
echo "2. 测试用户管理功能"
echo "3. 测试存储管理功能"
echo "4. 测试文件系统功能"
echo "5. 交互式模式"
echo ""

# Check if an argument is provided
if [ -n "$1" ]; then
    choice=$1
    echo "已选择操作: $choice"
else
    read -p "请选择操作 (1-5): " choice
fi

case $choice in
    1)
        echo "运行完整测试套件..."
        go run main.go test
        ;;
    2)
        echo "测试用户管理功能..."
        go run main.go test-user
        ;;
    3)
        echo "测试存储管理功能..."
        go run main.go test-storage
        ;;
    4)
        echo "测试文件系统功能..."
        go run main.go test-fs
        ;;
    5)
        echo "启动交互式模式..."
        go run main.go interactive
        ;;
    *)
        echo "默认启动交互式模式..."
        go run main.go
        ;;
esac

echo ""
echo "脚本执行完成"
