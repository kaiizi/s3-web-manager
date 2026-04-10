## 1. Project Scaffolding

- [x] 1.1 初始化 Go module：`backend/go.mod`（module name: `github.com/s3-web-manager/backend`）
- [x] 1.2 添加 Go 依赖：`gin-gonic/gin`、`aws-sdk-go-v2`（core/s3/presignedurl/config）、`golang-jwt/jwt/v5`、`joho/godotenv`
- [x] 1.3 创建 `backend/cmd/server/main.go`，完成 config 加载、storage provider 初始化、router 注册、服务启动
- [x] 1.4 创建 `backend/internal/config/config.go`，从环境变量读取所有配置（包括 `STORAGE_TYPE`），缺失必填项时 fatal 退出
- [x] 1.5 初始化 Vue 3 前端：`cd frontend && npm create vite@latest . -- --template vue-ts`
- [x] 1.6 安装前端依赖：`naive-ui`、`pinia`、`vue-router`、`axios`
- [x] 1.7 创建 `.env.example`，列出所有环境变量及说明（包括 `STORAGE_TYPE=ceph`）

## 2. Backend — Authentication

- [x] 2.1 创建 `backend/internal/api/auth.go`：实现 `POST /api/auth/login`，比对环境变量账密，成功返回 JWT
- [x] 2.2 创建 `backend/internal/middleware/auth.go`：JWT 验证中间件，验证失败返回 401
- [x] 2.3 在 `main.go` 注册路由：登录路由无需鉴权，其余路由组使用 JWT 中间件

## 3. Backend — Storage Abstraction Layer

- [x] 3.1 创建 `backend/internal/storage/provider.go`：定义 `StorageProvider` interface（`ListAllBuckets`、`GetBucketInfo`、`ListObjects`、`PresignUploadURL`、`PresignDownloadURL`、`DeleteObject`）及相关数据结构体（`BucketInfo`、`ObjectInfo`、`ListResult`）
- [x] 3.2 创建 `backend/internal/storage/ceph/provider.go`：定义 `CephProvider` struct，初始化 `aws-sdk-go-v2` S3 client（配置 endpoint、region、force-path-style）
- [x] 3.3 创建 `backend/internal/storage/ceph/admin.go`：实现 Ceph Admin OPS API HTTP 客户端，封装带 AWS v4 签名的 `ListAllBuckets` 和 `GetBucketInfo`（`stats=true`）技术方法
- [x] 3.4 在 `CephProvider` 中实现 `ListObjects`（ListObjectsV2, delimiter=/）、`PresignUploadURL`、`PresignDownloadURL`、`DeleteObject`，完成 `StorageProvider` 接口实现
- [x] 3.5 在 `main.go` 中根据 `STORAGE_TYPE` 环境变量通过 switch 初始化对应 provider，将 `StorageProvider` 实例注入各 API handler

## 4. Backend — Bucket API

- [x] 4.1 创建 `backend/internal/api/buckets.go`：实现 `GET /api/buckets`，调用 Admin OPS API 返回所有 bucket 列表（name + owner）
- [x] 4.2 实现 `GET /api/buckets/:name`，调用 Admin OPS API `stats=true`，返回 bucket 详情（name、owner、numObjects、sizeBytes）

## 5. Backend — Object API

- [x] 5.1 创建 `backend/internal/api/objects.go`：实现 `GET /api/buckets/:name/objects`，接受 `prefix` query 参数，使用 `ListObjectsV2`（delimiter=/）返回 `{prefixes, objects}`
- [x] 5.2 实现 `POST /api/buckets/:name/objects/upload-url`，接受 `{key, contentType}`，生成 Presigned PUT URL（15 分钟有效期）并返回
- [x] 5.3 实现 `GET /api/buckets/:name/objects/download-url`，接受 `key` query 参数，生成 Presigned GET URL（15 分钟有效期）并返回
- [x] 5.4 实现 `DELETE /api/buckets/:name/objects`，接受 `key` query 参数，调用 S3 `DeleteObject`，返回 204

## 6. Frontend — Router & Auth Store

- [x] 6.1 配置 `frontend/src/router/index.ts`：定义 `/login`、`/buckets`、`/buckets/:name` 路由，未登录时重定向到 `/login`
- [x] 6.2 创建 `frontend/src/stores/auth.ts`（Pinia）：管理 token 的存储（localStorage）、登录和登出动作
- [x] 6.3 配置 `frontend/src/api/` axios 实例：自动在请求头添加 `Authorization: Bearer <token>`，401 时跳转到登录页

## 7. Frontend — Login Page

- [x] 7.1 创建 `frontend/src/views/LoginView.vue`：用户名/密码表单，调用 `POST /api/auth/login`，成功后存 token 并跳转 `/buckets`
- [x] 7.2 表单验证：username 和 password 均不能为空

## 8. Frontend — Bucket List Page

- [x] 8.1 创建 `frontend/src/views/BucketListView.vue`：调用 `GET /api/buckets`，展示 Naive UI `NDataTable`（列：name、owner、操作）
- [x] 8.2 实现点击 bucket 行跳转到 `/buckets/:name`

## 9. Frontend — Object Browser Page

- [x] 9.1 创建 `frontend/src/views/BucketDetailView.vue`：展示当前 prefix 的 breadcrumb 导航，调用 `GET /api/buckets/:name/objects?prefix=<current>`
- [x] 9.2 创建 `frontend/src/components/ObjectTable.vue`：展示 prefixes（可点击进入）和 objects（显示 key、大小、最后修改时间），每行有下载和删除按钮
- [x] 9.3 实现点击 prefix 更新 currentPrefix 状态并重新拉取数据
- [x] 9.4 实现下载：调用获取 Presigned URL 接口，通过 `window.location.href` 触发浏览器下载
- [x] 9.5 实现删除：点击删除按钮弹出确认对话框（`NModal`），确认后调用 `DELETE /api/buckets/:name/objects?key=<key>`，刷新列表

## 10. Frontend — Upload

- [x] 10.1 创建 `frontend/src/components/UploadModal.vue`：文件选择 +「上传」按钮，调用后端获取 Presigned PUT URL，再用 XMLHttpRequest 直接 PUT 到 Ceph（支持进度条）
- [x] 10.2 在 `BucketDetailView.vue` 中集成上传按钒，上传完成后刷新 object 列表
- [x] 10.3 上传时在对话框内显示进度百分比（`NProgress`）

## 11. Project Engineering

- [x] 11.1 创建 `backend/Dockerfile`：multi-stage 构建（golang:1.23-alpine 构建 Go 二进制 → alpine 最终镜像，指定工作目录、复制二进制、EXPOSE 8080）
- [x] 11.2 创建 `frontend/nginx.conf`：配置 SPA fallback（`try_files $uri $uri/ /index.html`）和 `/api` 反代到 `backend:8080`
- [x] 11.3 创建 `frontend/Dockerfile`：multi-stage 构建（node:20-alpine 运行 `npm run build` → nginx:alpine 将 dist 复制到 `/usr/share/nginx/html`，包含自定义 nginx.conf）
- [x] 11.4 创建 `docker-compose.yml`：定义 `backend`（构建 `./backend`、env_file: .env、暑露 8080）和 `frontend`（构建 `./frontend`、暕露 80、depends_on backend）两个 service
- [x] 11.5 配置 `frontend/vite.config.ts`：开发模式下 `/api` 代理到 `http://localhost:8080`，生产构建不包含代理配置
- [x] 11.6 创建 `README.md`：项目介绍、前置要求、docker-compose 快速启动步骤、环境变量说明表、Ceph CORS 配置命令示例、本地开发步骤
