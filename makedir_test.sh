# 登录并提取 token（假设 jq 可用）
TOKEN=$(curl -s -X POST "http://localhost:8080/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"suzuki","password":"suzuki"}' | jq -r '.data.token')

# 用 token 创建目录
curl -v -X POST "http://localhost:8080/api/fs/mkdir" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer '"$TOKEN"'" \
  -d '{"path":"/localwebdav/newdir"}'