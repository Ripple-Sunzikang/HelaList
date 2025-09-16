# ./makedir_test.sh <username> <password> <dirname>
set -euo pipefail

if [ $# -lt 3 ]; then
    echo "用法: $0 <username> <password> <dirname>"
    exit 1
fi

USERNAME="$1"
PASSWORD="$2"
DIRNAME="$3"

LOGIN_URL="http://localhost:8080/api/user/login"
MKDIR_URL="http://localhost:8080/api/fs/mkdir"

# 登录并提取 token（需要 jq）
TOKEN=$(curl -s -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo "登录失败或未获取到 token"
    exit 1
fi

# 用 token 创建目录
curl -v -X POST "$MKDIR_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"path\":\"/localwebdav/$DIRNAME\"}"