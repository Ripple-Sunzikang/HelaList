#!/bin/bash
# ./rename_test.sh <username> <password> <source_path> <new_name>
set -euo pipefail

if [ $# -lt 4 ]; then
    echo "用法: $0 <username> <password> <source_path> <new_name>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
SOURCE_PATH="$3"
NEW_NAME="$4"

LOGIN_URL="http://localhost:8080/api/user/login"
RENAME_URL="http://localhost:8080/api/fs/rename"

# 登录并提取 token（需要 jq）
echo "正在登录并获取 Token..."
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi
echo "Token 获取成功。"
echo ""

# 使用 token 重命名文件/目录
echo "正在发送重命名请求..."
curl -v -X POST "$RENAME_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"path\":\"$SOURCE_PATH\",\"name\":\"$NEW_NAME\"}"