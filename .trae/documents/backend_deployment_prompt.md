# OpenClaw 后端本地部署提示词

## 目标
在本地环境中成功部署和运行 OpenClaw 后端服务，确保前端能够正常访问。

## 前提条件

### 必须安装的软件
1. **Git** - 版本控制工具
2. **Go** - 1.20+ 版本（后端开发语言）
3. **PostgreSQL** - 14+ 版本（唯一支持的数据库）
4. **Docker** 和 **Docker Compose** - 推荐的本地开发方式
5. **Node.js** 和 **npm** - 前端开发和构建

## 部署步骤

### 方案 1: 使用 Docker Compose（推荐）

#### 1. 克隆代码仓库
```bash
git clone <repository-url>
cd <repository-directory>
```

#### 2. 配置环境变量
进入本地开发栈目录：
```bash
cd .temp/dev-stack
```

复制环境变量示例文件并填写：
```bash
cp .env.example .env
```

编辑 `.env` 文件，设置以下关键参数：
```
# PostgreSQL 配置
POSTGRES_DB=openclaw
POSTGRES_USER=openclaw
POSTGRES_PASSWORD=openclaw123
POSTGRES_PORT=5432

# Redis 配置
REDIS_PORT=6379

# MinIO 配置
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin123
MINIO_API_PORT=9000
MINIO_CONSOLE_PORT=9001
MINIO_BUCKET=openclaw

# OpenSearch 配置
OPENSEARCH_PORT=9200
OPENSEARCH_DASHBOARDS_PORT=5601

# 平台 API 配置
PLATFORM_API_PORT=8080
```

#### 3. 启动开发栈
```bash
docker-compose up -d
```

这会启动以下服务：
- PostgreSQL 数据库
- Redis 缓存
- MinIO 对象存储
- OpenSearch 搜索引擎

#### 4. 初始化数据库
等待服务启动完成后，执行数据库引导：
```bash
# 进入后端目录
cd ../../apps/server

# 设置环境变量并执行引导
DATABASE_URL=postgresql://openclaw:openclaw123@localhost:5432/openclaw \
BOOTSTRAP_DATA_PATH=./bootstrap-data.json \
go run ./cmd/bootstrap
```

如果没有 `bootstrap-data.json` 文件，可以创建一个基本的：
```json
{
  "tenants": [],
  "users": [],
  "instances": [],
  "channels": []
}
```

#### 5. 启动后端服务
```bash
# 设置环境变量
DATABASE_URL=postgresql://openclaw:openclaw123@localhost:5432/openclaw \
PLATFORM_STRICT_MODE=true \

# 启动服务
go run ./cmd/server
```

### 方案 2: 手动部署（不使用 Docker）

#### 1. 安装 PostgreSQL
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql-server
sudo postgresql-setup --initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql

# macOS
brew install postgresql
brew services start postgresql
```

#### 2. 创建数据库和用户
```bash
# 登录 PostgreSQL
sudo -u postgres psql

# 创建数据库和用户
CREATE DATABASE openclaw;
CREATE USER openclaw WITH PASSWORD 'openclaw123';
GRANT ALL PRIVILEGES ON DATABASE openclaw TO openclaw;
\q
```

#### 3. 配置环境变量
在 `apps/server` 目录创建 `.env` 文件：
```bash
cd apps/server
cp .env.example .env
```

编辑 `.env` 文件：
```
DATABASE_URL=postgresql://openclaw:openclaw123@localhost:5432/openclaw
PLATFORM_STRICT_MODE=true
```

#### 4. 初始化数据库
```bash
go run ./cmd/bootstrap
```

#### 5. 启动后端服务
```bash
go run ./cmd/server
```

## 验证步骤

### 检查服务是否运行
1. 后端服务默认运行在 `http://localhost:8080`
2. 测试健康检查端点：
   ```bash
   curl http://localhost:8080/healthz
   ```
   预期返回：`ok`

3. 测试版本端点：
   ```bash
   curl http://localhost:8080/versionz
   ```
   预期返回版本信息

### 启动前端服务
```bash
cd apps/web
npm install
npm run dev
```

前端服务默认运行在 `http://localhost:5173`

## 常见问题排查

### 数据库连接失败
- 检查 PostgreSQL 是否运行：`sudo systemctl status postgresql`
- 检查数据库连接字符串是否正确
- 检查防火墙是否允许 5432 端口

### 后端启动失败
- 检查 `DATABASE_URL` 环境变量是否设置
- 检查 PostgreSQL 是否正常运行
- 查看控制台错误信息

### 前端无法访问后端
- 检查后端服务是否运行在 8080 端口
- 检查前端 `VITE_DEV_API_PROXY_TARGET` 配置是否正确
- 尝试直接访问后端 API 端点

## 生产环境注意事项
- 不要使用默认的密码和配置
- 启用 HTTPS
- 配置适当的防火墙规则
- 定期备份数据库
- 使用环境变量管理敏感信息

## 成功标志
- 后端服务能够正常启动
- 前端能够成功登录并进入 portal 页面
- API 端点能够正常响应请求

现在您应该可以在本地环境中成功运行 OpenClaw 后端服务了！