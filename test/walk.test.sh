# 用法: ./walk.test.sh <username> <password> [path] [depth]
# 默认 path=localwebdav, depth=3
set -euo pipefail
if [ $# -lt 2 ] || [ $# -gt 4 ]; then
    echo "用法: $0 <username> <password> [path] [depth]"
    exit 1
fi

USERNAME=$1
PASSWORD=$2
PATH_ARG=${3:-localwebdav}
DEPTH=${4:-3}

LOGIN_URL="http://localhost:8080/api/user/login"
WALK_URL="http://localhost:8080/api/fs/walk/${PATH_ARG}?depth=${DEPTH}"

# 登录获取 token
LOGIN_RESPONSE=$(curl -s -X POST "$LOGIN_URL" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
    echo "登录成功，token: $TOKEN"
else
    echo "登录失败："
    echo "$LOGIN_RESPONSE" | jq '.'
    exit 1
fi

# 发起遍历请求
WALK_RESPONSE=$(curl -s -X GET "$WALK_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN")

# 输出结果
echo "walk response:"
echo "$WALK_RESPONSE" | jq '.'