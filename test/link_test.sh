#!/bin/bash
# ./link_test.sh <username> <password> <file_path>
# 示例: ./link_test.sh user1 123456 /localwebdav/file.txt
set -euo pipefail

if [ $# -lt 3 ]; then
    echo "用法: $0 <username> <password> <file_path>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
FILEPATH="$3"

LOGIN_URL="http://localhost:8080/api/user/login"
LINK_URL="http://localhost:8080/api/fs/link"

# 登录并提取 token（需要 jq）
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi

echo "使用 Token: $TOKEN"

# 使用 token 获取文件链接
curl -v -X POST "$LINK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"path\":\"$FILEPATH\"}"