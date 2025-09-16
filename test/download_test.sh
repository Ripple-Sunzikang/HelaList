#!/bin/bash
# ./download_test.sh <username> <password> <file_path>
# 示例: ./download_test.sh user1 123456 /localwebdav/file.txt
set -euo pipefail

if [ $# -lt 3 ]; then
    echo "用法: $0 <username> <password> <file_path>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
FILEPATH="$3"
# 从文件路径中提取文件名作为本地保存的文件名
FILENAME=$(basename "$FILEPATH")

LOGIN_URL="http://localhost:8080/api/user/login"
LINK_URL="http://localhost:8080/api/fs/link"

# 1. 登录并提取 token
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi

echo "使用 Token 成功登录"

# 2. 获取文件链接URL
echo "正在获取文件 '$FILEPATH' 的下载链接..."
DOWNLOAD_URL=$(curl -s -X POST "$LINK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"path\":\"$FILEPATH\"}" | jq -r '.data.url // empty')

if [ -z "$DOWNLOAD_URL" ]; then
    echo "获取下载链接失败"
    exit 1
fi

echo "获取到下载链接: $DOWNLOAD_URL"

# 3. 使用获取到的URL下载文件
echo "正在下载文件到: $FILENAME ..."
# 使用 curl 的 -L 选项来自动处理重定向
# 使用 curl 的 -o 选项来指定输出文件名
curl -L -o "$FILENAME" "$DOWNLOAD_URL"

echo "文件 '$FILENAME' 下载完成."