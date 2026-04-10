## Why

当前缺乏一个统一的 Web 界面来管理 Ceph S3 存储，运维人员只能通过 CLI 工具（awscurl、s3cmd 等）操作，效率低下且门槛高。需要一个具备管理员视角的 Web 平台，能够跨用户查看所有 bucket 并进行 object 的上传、下载和删除操作。

## What Changes

- 新建 Go 后端服务，提供 REST API，集成 Ceph RGW Admin OPS API 与标准 S3 API
- 新建 Vue 3 独立前端，提供现代化 Web 界面，支持 bucket 浏览与 object 管理
- 实现单用户固定账密认证，JWT 令牌保护所有 API
- Object 上传/下载采用 Presigned URL 直连 Ceph，后端不做数据中转
- 提供完整的项目工程化文件：Dockerfile（multi-stage）、docker-compose、README

## Capabilities

### New Capabilities

- `authentication`: 单用户固定账密登录，JWT 签发与验证，保护所有后端 API
- `bucket-management`: 通过 Ceph Admin OPS API 列出所有 bucket（跨 owner），查看 bucket 详情与统计信息
- `object-management`: 列出 bucket 内 objects（支持 prefix 模拟目录树），上传、下载、删除 object，上传/下载通过 Presigned URL 直连 Ceph
- `project-engineering`: Dockerfile multi-stage 构建、docker-compose 编排、README 文档、.env.example 配置模板

### Modified Capabilities

（无，全新项目）

## Impact

- **新增依赖**：Go 后端（aws-sdk-go-v2、gin/echo、golang-jwt）；前端（Vue 3、Vite、Naive UI、axios）
- **外部依赖**：Ceph RGW v18.4，需要 Admin caps 凭证（buckets/users/metadata/usage/zone=*）
- **网络要求**：浏览器可直接访问 Ceph RGW endpoint（Presigned URL 直连所必需）
- **部署方式**：单个 Docker 镜像包含前后端，通过 docker-compose 一键启动
