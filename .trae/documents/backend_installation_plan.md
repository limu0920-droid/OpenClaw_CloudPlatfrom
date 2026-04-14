# 后端服务安装计划

## 环境现状分析

### 已安装组件
- Go 1.25.1

### 缺失组件
- PostgreSQL 数据库
- 其他可能需要的服务：Redis、MinIO（根据项目配置）

## 安装计划

### 步骤 1: 安装 PostgreSQL
1. 更新系统包管理器
2. 安装 PostgreSQL
3. 启动 PostgreSQL 服务
4. 创建数据库和用户

### 步骤 2: 配置 PostgreSQL
1. 连接到 PostgreSQL
2. 创建 `platform` 数据库
3. 创建 `platform` 用户并设置密码
4. 授予用户对数据库的权限

### 步骤 3: 配置后端环境
1. 更新 [.env](file:///workspace/apps/server/.env) 文件，设置正确的数据库连接字符串
2. 配置其他必要的环境变量

### 步骤 4: 初始化数据库
1. 运行数据库迁移
2. 执行数据库引导（如果需要）

### 步骤 5: 启动后端服务
1. 启动 Go 后端服务
2. 验证服务是否正常运行

## 详细安装命令

### PostgreSQL 安装
```bash
# 更新系统包管理器
sudo apt update

# 安装 PostgreSQL
sudo apt install -y postgresql postgresql-contrib

# 启动 PostgreSQL 服务
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### PostgreSQL 配置
```bash
# 连接到 PostgreSQL 并创建数据库和用户
sudo -u postgres psql -c "CREATE DATABASE platform;"
sudo -u postgres psql -c "CREATE USER platform WITH PASSWORD 'platform-dev-password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE platform TO platform;"
sudo -u postgres psql -c "ALTER USER platform WITH SUPERUSER;"
```

### 后端环境配置
更新 [.env](file:///workspace/apps/server/.env) 文件：
```
DATABASE_URL=postgresql://platform:platform-dev-password@localhost:5432/platform?sslmode=disable
AUTO_MIGRATE=true
```

### 启动后端服务
```bash
cd /workspace/apps/server
go run ./cmd/server
```

## 验证步骤
1. 检查后端服务是否成功启动
2. 测试 API 端点，如 `/healthz`
3. 验证前端是否能与后端正常通信

## 潜在风险与解决方案

### 风险 1: PostgreSQL 安装失败
- **解决方案**: 检查系统包管理器状态，尝试使用不同的安装源

### 风险 2: 数据库连接失败
- **解决方案**: 检查数据库服务状态、用户权限和连接字符串

### 风险 3: 后端服务启动失败
- **解决方案**: 检查环境变量配置，查看服务日志以定位错误

## 依赖关系
- Go 1.25.1 (已安装)
- PostgreSQL 16+
- 网络连接（用于下载依赖）
