# S3 Web Manager

一个基于 Go + Vue 3 的 S3 存储 Web 管理平台，专为 Ceph RGW 设计，支持管理员视角查看所有 Bucket 并进行 Object 的上传、下载和删除操作。

## 功能特性

- **管理员视角**：通过 Ceph Admin OPS API 查看所有 Bucket（不受 owner 限制）
- **Object 浏览**：prefix 模拟目录树导航，支持多级目录
- **文件操作**：上传（Presigned URL 直连 Ceph）、下载、删除
- **JWT 认证**：固定账密登录，24 小时令牌有效期
- **前后端分离**：独立容器，docker-compose 一键启动
- **可扩展**：StorageProvider 抽象层，便于后续支持 MinIO 等

## 前置要求

- Docker >= 24 + Docker Compose v2
- Ceph RGW v18.4，Admin 凭证（需要 caps: `buckets=*;users=*;usage=*;metadata=*`）
- 浏览器能直接访问 Ceph RGW endpoint（Presigned URL 直连所必需）

## 快速启动

```bash
# 1. 复制配置文件
cp .env.example .env

# 2. 编辑 .env，填写实际配置
vim .env

# 3. 构建镜像并启动服务
docker-compose up -d --build

# 4. 访问 Web 界面
# http://localhost
```

## 打包镜像

### 使用 docker-compose 构建（推荐）

```bash
# 构建所有镜像
docker-compose build

# 构建指定服务
docker-compose build backend
docker-compose build frontend
```

### 单独构建镜像

```bash
# 构建后端镜像
docker build -t s3-web-manager-backend:latest ./backend

# 构建前端镜像
docker build -t s3-web-manager-frontend:latest ./frontend
```

### 推送到镜像仓库

```bash
# 打标签
docker tag s3-web-manager-backend:latest <your-registry>/s3-web-manager-backend:latest
docker tag s3-web-manager-frontend:latest <your-registry>/s3-web-manager-frontend:latest

# 推送
docker push <your-registry>/s3-web-manager-backend:latest
docker push <your-registry>/s3-web-manager-frontend:latest
```

## 环境变量说明

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `APP_USERNAME` | ✅ | — | 登录用户名 |
| `APP_PASSWORD` | ✅ | — | 登录密码 |
| `JWT_SECRET` | ✅ | — | JWT 签名密钥（至少 32 个字符） |
| `APP_PORT` | | `8080` | 后端监听端口 |
| `STORAGE_TYPE` | | `ceph` | 存储类型（当前支持 `ceph`） |
| `CEPH_ENDPOINT` | ✅ | — | Ceph RGW 地址，例如 `http://172.16.31.61:8000` |
| `CEPH_ACCESS_KEY` | ✅ | — | Admin 访问密钥 |
| `CEPH_SECRET_KEY` | ✅ | — | Admin 私有密钥 |
| `CEPH_REGION` | | `default` | S3 Region |

> **注意**：`CEPH_ENDPOINT` 必须是浏览器可直接访问的地址，因为上传/下载使用 Presigned URL 直连 Ceph。

## Ceph CORS 配置

浏览器通过 Presigned URL 直接向 Ceph 上传文件时，Ceph 需要允许对应的 Origin。对每个需要操作的 Bucket 执行：

```bash
# 安装 s3cmd
pip install s3cmd

# 配置 s3cmd（使用 admin 凭证）
s3cmd --configure

# 创建 cors.xml
cat > /tmp/cors.xml << 'EOF'
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>http://your-frontend-host</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedMethod>HEAD</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
    <MaxAgeSeconds>3000</MaxAgeSeconds>
  </CORSRule>
</CORSConfiguration>
EOF

# 为 bucket 设置 CORS
s3cmd setcors /tmp/cors.xml s3://your-bucket-name

# 或使用 awscurl
awscurl \
  --access_key <AK> \
  --secret_key <SK> \
  --region default \
  --service s3 \
  -X PUT \
  --data-binary @/tmp/cors.xml \
  "http://172.16.31.61:8000/your-bucket-name?cors"
```

也可以设置通配符 `AllowedOrigin` 为 `*`（仅限内网环境）。

## 本地开发

### 后端

```bash
cd backend
cp ../.env.example .env      # 修改为本地配置
go run ./cmd/server
# 后端运行在 http://localhost:8080
```

### 前端

```bash
cd frontend
npm install
npm run dev
# 前端运行在 http://localhost:5173
# /api 请求自动代理到 http://localhost:8080
```

## 项目结构

```
s3-web-manager/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── api/                    # HTTP handlers
│   │   ├── config/                 # 环境变量配置
│   │   ├── middleware/auth.go      # JWT 中间件
│   │   └── storage/
│   │       ├── provider.go         # StorageProvider interface
│   │       └── ceph/               # Ceph 实现
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── views/                  # 页面组件
│   │   ├── components/             # 复用组件
│   │   ├── api/                    # axios 封装
│   │   ├── stores/auth.ts          # Pinia token 管理
│   │   └── router/index.ts         # Vue Router
│   ├── nginx.conf                  # nginx 配置（SPA + API 反代）
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```
