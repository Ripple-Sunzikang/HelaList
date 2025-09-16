#!/bin/bash
# ./move_test.sh <username> <password> <source_path> <destination_path>
# 示例: ./move_test.sh user1 123456 /localwebdav/file.txt /localwebdav/new_folder/file_moved.txt
set -euo pipefail

if [ $# -lt 4 ]; then
    echo "用法: $0 <username> <password> <source_path> <destination_path>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
SRCPATH="$3"
DSTPATH="$4"

LOGIN_URL="http://localhost:8080/api/user/login"
MOVE_URL="http://localhost:8080/api/fs/move"

# 登录并提取 token（需要 jq）
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi

echo "使用 Token: $TOKEN"

# 使用 token 移动文件/目录
curl -v -X POST "$MOVE_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"src_path\":\"$SRCPATH\",\"dst_path\":\"$DSTPATH\"}"