#!/bin/bash
# ./copy_test.sh <username> <password> <source_path> <destination_path>
# 示例: ./copy_test.sh user1 123456 /localwebdav/file.txt /localwebdav/file_copy.txt
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
COPY_URL="http://localhost:8080/api/fs/copy"

# 登录并提取 token（需要 jq）
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi

echo "使用 Token: $TOKEN"

# 使用 token 复制文件/目录
curl -v -X POST "$COPY_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"src_path\":\"$SRCPATH\",\"dst_path\":\"$DSTPATH\"}"