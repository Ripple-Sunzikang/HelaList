# ./query_test.sh <username> <password>
set -euo pipefail
if [ $# -ne 2 ]; then
    echo "用法: $0 <username> <password>"
    exit 1
fi

USERNAME=$1
PASSWORD=$2
LOGIN_URL="http://localhost:8080/api/user/login"
LIST_URL="http://localhost:8080/api/fs/list/localwebdav?refresh=true"

# 登录获取token
LOGIN_RESPONSE=$(curl -s -X POST "$LOGIN_URL" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# 检查登录是否成功
if echo "$LOGIN_RESPONSE" | jq -e '.code == 200' > /dev/null; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
    echo "登录成功，获取到token: $TOKEN"
    
    # 查询文件列表
    LIST_RESPONSE=$(curl -s -X GET "$LIST_URL" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN")
    
    # 输出结果
    echo "$LIST_RESPONSE" | jq '.'
else
    echo "登录失败："
    echo "$LOGIN_RESPONSE" | jq '.'
fi