#!/bin/bash
# 用法: ./put_test.sh <username> <password> <local_file_to_upload> <remote_dir_path>
# 示例: ./put_test.sh admin admin ./test.txt /localwebdav

set -euo pipefail

if [ $# -lt 4 ]; then
    echo "用法: $0 <username> <password> <local_file_to_upload> <remote_dir_path>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
LOCAL_FILE="$3"
REMOTE_DIR="$4"

LOGIN_URL="http://localhost:8080/api/user/login"
PUT_URL="http://localhost:8080/api/fs/put"

# 检查本地文件是否存在
if [ ! -f "$LOCAL_FILE" ]; then
    echo "错误: 本地文件 '$LOCAL_FILE' 不存在。"
    # 你也可以在这里创建一个临时文件用于测试
    # echo "创建一个临时测试文件..."
    # echo "这是 HelaList 的一个测试文件。" > "$LOCAL_FILE"
    exit 1
fi

# 登录并提取 token（需要 jq）
echo "正在登录..."
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi
echo "登录成功, Token: $TOKEN"
echo "-----------------------------------"

# 使用 token 上传文件
# -F 用于发送 multipart/form-data
# -F "path=$REMOTE_DIR" 发送目标路径
# -F "file=@$LOCAL_FILE" 发送文件内容
echo "正在上传文件 '$LOCAL_FILE' 到 '$REMOTE_DIR'..."
curl -v -X POST "$PUT_URL" \
  -H "Authorization: Bearer $TOKEN" \
  -F "path=$REMOTE_DIR" \
  -F "file=@$LOCAL_FILE"

echo -e "\n\n上传请求已发送。"