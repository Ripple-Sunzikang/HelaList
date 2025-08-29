#!/bin/bash
# ./remove_test.sh <username> <password> <path_to_delete>
set -euo pipefail

if [ $# -lt 3 ]; then
    echo "用法: $0 <username> <password> <path_to_delete>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
TARGET_PATH="$3"

LOGIN_URL="http://localhost:8080/api/user/login"
REMOVE_URL="http://localhost:8080/api/fs/remove"

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

# 使用 token 删除文件/目录
echo "正在发送删除请求..."
curl -v -X POST "$REMOVE_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"path\":\"$TARGET_PATH\"}"