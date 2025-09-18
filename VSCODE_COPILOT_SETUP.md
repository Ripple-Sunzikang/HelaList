# VSCode Copilot + DBHub PostgreSQL 配置说明

## 已完成的配置

### 1. MCP 服务器配置
- 位置: `~/.config/Code/User/globalStorage/github.copilot-chat/mcp_servers.json`
- 配置了名为 `dbhub-hela` 的MCP服务器
- 使用只读模式连接到你的PostgreSQL数据库

### 2. 数据库连接信息
- 主机: localhost
- 端口: 5432  
- 用户: suzuki
- 密码: suzuki
- 数据库: hela
- SSL模式: 禁用
- **只读模式**: 已启用（安全）

### 3. 环境配置文件
- 创建了 `.env.dbhub` 文件，包含数据库连接配置

## 如何在VSCode中使用

### 启用步骤

1. **重启VSCode**: 让新的MCP配置生效
2. **打开Copilot Chat**: 使用 `Ctrl+Shift+P` 然后搜索 "Copilot Chat"
3. **验证连接**: 在Chat中询问关于数据库的问题

### 可用功能

#### 📊 数据库探索
```
请显示 hela 数据库中的所有表
```

#### 🔍 查看表结构  
```
请显示 users 表的结构
```

#### 📝 生成SQL查询
```
帮我写一个查询，获取所有用户的姓名和邮箱
```

#### ⚡ 执行SQL（只读）
```
执行这个查询: SELECT * FROM users LIMIT 10;
```

#### 🧠 解释数据库元素
```
解释一下这个数据库的schema设计
```

### 安全特性

- ✅ **只读模式**: 只能执行SELECT查询，无法修改数据
- ✅ **连接安全**: 使用SSL禁用模式连接本地数据库
- ✅ **权限控制**: 基于数据库用户权限

### 故障排除

如果遇到问题：

1. **检查PostgreSQL服务**:
   ```bash
   sudo systemctl status postgresql
   ```

2. **测试连接**:
   ```bash
   psql -h localhost -p 5432 -U suzuki -d hela
   ```

3. **手动测试DBHub**:
   ```bash
   export DSN="postgres://suzuki:suzuki@localhost:5432/hela?sslmode=disable"
   npx @bytebase/dbhub --readonly
   ```

4. **查看VSCode输出**: 在VSCode中打开"输出"面板，选择"GitHub Copilot Chat"

### 配置文件位置

- VSCode MCP配置: `~/.config/Code/User/globalStorage/github.copilot-chat/mcp_servers.json`
- 环境变量文件: `~/codes/HelaList/.env.dbhub`

---

**注意**: 如果需要修改数据库连接信息，请编辑上述配置文件并重启VSCode。