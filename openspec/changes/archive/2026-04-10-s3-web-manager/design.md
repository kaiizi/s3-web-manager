## Context

全新项目，无现有代码基础。目标是构建一个内部运维工具，提供 Web 界面管理 Ceph v18.4 (Reef) RGW 的 S3 存储。使用者为单一管理员，需要跨 owner 查看所有 bucket 并进行 object 的增删查操作。

技术约束：
- 后端：Go 语言
- S3 平台：Ceph RGW v18.4，需保留对其他 S3 兼容接口的扩展性
- 前端：独立现代前端（Vue 3 + Vite）
- 部署：Docker / docker-compose
- 浏览器可直接访问 Ceph RGW endpoint

## Goals / Non-Goals

**Goals:**
- 管理员视角列出所有 bucket（Ceph Admin OPS API）
- 查看 bucket 详情及 object 列表（prefix 模拟目录树）
- Object 上传、下载、删除
- Presigned URL 直连 Ceph（后端不做数据中转）
- 单用户固定账密认证，JWT 保护
- 前后端分离容器，docker-compose 一键部署，通过 .env 文件注入环境变量
- 后端 S3 客户端具备抽象层，当前实现 Ceph provider，预留 MinIO 等扩展点
- 无外部中间件依赖（无 Redis、消息队列等），所有状态内存或本地文件持久化

**Non-Goals:**
- 多用户/账号体系
- Bucket 创建/删除
- Object 重命名（copy + delete 复杂度不纳入 v1）
- ACL / Policy 管理
- Ceph 集群监控
- 移动端适配

## Decisions

### D1. 列出所有 Bucket — 使用 Ceph Admin OPS API

**决策**：通过 `GET /admin/bucket?list` 列出全局所有 bucket，用 `GET /admin/bucket?bucket=<name>&stats=true` 获取详情。

**背景**：标准 S3 `ListAllMyBuckets` 只返回当前凭证 owner 的 bucket；Ceph Admin OPS API 无此限制。

**替代方案**：
- 标准 S3 API：❌ 无法跨 owner 看到所有 bucket
- Ceph Admin Python SDK：❌ 需要额外运行时，Go 直接 HTTP 调用更简洁

**实现**：Go 后端持有 admin AK/SK，调用 Admin OPS API 时使用 AWS v4 签名（服务名 `s3`，host 为 RGW endpoint）。

---

### D2. Object 上传/下载 — Presigned URL 直连

**决策**：后端生成 Presigned URL，前端直接与 Ceph 通信，Go 不做数据流量中转。

**理由**：
- 避免大文件占用 Go 服务内存/带宽
- 浏览器可直接访问 Ceph RGW（已确认），此方案可行
- 与 MinIO Console 的交互模式一致

**上传流程**：
```
前端 → POST /api/buckets/:name/objects/upload-url {key, contentType}
     ← {presignedUrl, method: "PUT"}
前端 → PUT <presignedUrl> (直接到 Ceph)
```

**下载流程**：
```
前端 → GET /api/buckets/:name/objects/download-url?key=<key>
     ← {presignedUrl}
前端 → location.href = presignedUrl （触发浏览器下载）
```

**替代方案**：
- Go 代理中转（`io.Copy`）：❌ 大文件内存压力，不选
- 分片上传（Multipart）：⏸ v1 不实现，大文件场景 v2 扩展

---

### D3. 前端技术栈 — Vue 3 + Vite + Naive UI

**决策**：Vue 3 (Composition API) + Vite + Naive UI + Pinia + Vue Router + axios

**理由**：
- Naive UI 有成熟的 Upload 组件，支持进度显示
- Vite 构建产物为纯静态文件，Docker 友好
- Composition API 适合文件管理类状态逻辑

**替代方案**：
- React + Ant Design：功能等价，无决定性优势
- Go `html/template` SSR：❌ UI 难以做精致

---

### D4. 后端框架 — Gin

**决策**：使用 `gin-gonic/gin` 作为 HTTP 框架。

**理由**：轻量、路由清晰、中间件生态成熟（JWT middleware 易接入）。

**替代方案**：Echo —— 等价，Gin 社区更大。

---

### D5. 认证 — JWT HS256，固定账密从环境变量读取

**决策**：登录时比对环境变量 `APP_USERNAME` / `APP_PASSWORD`（明文，内部工具可接受），成功后签发 JWT（HS256，24h 有效期），后续请求通过 `Authorization: Bearer <token>` 验证。

**理由**：单用户场景无需数据库，环境变量配置简单，符合 Docker 部署习惯。

**安全考虑**：
- JWT secret 从环境变量 `JWT_SECRET` 读取（≥32 字符随机串）
- HTTPS 由上层反代负责（内部工具，docker-compose 不强制）

---

### D6. Docker 部署 — 前后端分离容器，docker-compose 编排

**决策**：  
- `frontend/Dockerfile`：multi-stage（Node 构建 → nginx:alpine 托管静态文件）  
- `backend/Dockerfile`：multi-stage（Go 构建 → alpine 运行二进制）  
- `docker-compose.yml`：定义 `backend` 和 `frontend` 两个 service，通过 `env_file: .env` 注入配置，frontend nginx 反代 `/api` 到 backend

**理由**：
- 前后端独立迭代，镜像职责清晰
- nginx 反代 API 解决浏览器跨域问题，无需额外 CORS 头处理
- 环境变量通过 `.env` 文件统一管理，docker-compose 一键启动
- 无需 embed.FS 或文件挂载，构建产物路径明确

**nginx 路由规则**（frontend container）：
```
location /api {  proxy_pass http://backend:8080; }
location /     {  try_files $uri $uri/ /index.html; }  # SPA fallback
```

**替代方案**：
- 单镜像 embed.FS：前后端耦合，更新前端需重新构建 Go binary，不选
- 挂载静态文件目录：容器内路径依赖外部卷，不优雅

---

### D7. 项目目录结构

```
s3-web-manager/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── api/
│   │   │   ├── auth.go          # POST /api/auth/login
│   │   │   ├── buckets.go       # GET /api/buckets, /api/buckets/:name
│   │   │   └── objects.go       # 列举、presign-upload、presign-download、删除
│   │   ├── storage/
│   │   │   ├── provider.go      # StorageProvider interface（抽象层）
│   │   │   └── ceph/
│   │   │       ├── provider.go  # CephProvider 实现（标准 S3 操作）
│   │   │       └── admin.go     # Ceph Admin OPS API HTTP 客户端
│   │   ├── middleware/
│   │   │   └── auth.go          # JWT 验证中间件（Gin handler，非外部服务）
│   │   └── config/
│   │       └── config.go        # 环境变量加载
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── views/
│   │   │   ├── LoginView.vue
│   │   │   ├── BucketListView.vue
│   │   │   └── BucketDetailView.vue
│   │   ├── components/
│   │   │   ├── ObjectTable.vue      # object 列表 + 操作按钮
│   │   │   └── UploadModal.vue      # 上传对话框（进度显示）
│   │   ├── api/
│   │   │   ├── auth.ts
│   │   │   ├── buckets.ts
│   │   │   └── objects.ts
│   │   ├── stores/
│   │   │   └── auth.ts              # Pinia: token 存储
│   │   ├── router/index.ts
│   │   └── main.ts
│   ├── nginx.conf               # SPA fallback + /api 反代配置
│   ├── Dockerfile
│   ├── package.json
│   └── vite.config.ts
│
├── docker-compose.yml
├── .env.example
└── README.md
```

---

### D8. S3 客户端抽象层 — StorageProvider interface

**决策**：在 `backend/internal/storage/provider.go` 定义 `StorageProvider` interface，`CephProvider` 实现该接口，未来 MinIO、AWS S3 等只需新增 provider 包。

**接口设计**：
```go
type BucketInfo struct {
    Name       string
    Owner      string
    NumObjects int64
    SizeBytes  int64
}

type ListResult struct {
    Prefixes []string
    Objects  []ObjectInfo
}

type ObjectInfo struct {
    Key          string
    Size         int64
    LastModified time.Time
}

type StorageProvider interface {
    // Admin-level: returns all buckets regardless of owner
    ListAllBuckets(ctx context.Context) ([]BucketInfo, error)
    GetBucketInfo(ctx context.Context, bucket string) (BucketInfo, error)

    // Object operations
    ListObjects(ctx context.Context, bucket, prefix string) (*ListResult, error)
    PresignUploadURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error)
    PresignDownloadURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
    DeleteObject(ctx context.Context, bucket, key string) error
}
```

**扩展路径**：
- `storage/ceph/provider.go` — 当前实现（Admin OPS API + aws-sdk-go-v2）
- `storage/minio/provider.go` — 未来（MinIO Go SDK `mcapi`）
- `storage/s3/provider.go` — 未来（原生 AWS S3，`ListBuckets` 标准接口）

引导实例化由 `config.StorageType`（环境变量 `STORAGE_TYPE`，默认 `ceph`）决定，通过简单 switch 而非注册表完成。

---

### D9. 无外部中间件依赖

**决策**：系统不依赖任何外部中间件服务（无 Redis、消息队列、外部 session store 等）。

**理由**：内部单用户工具，JWT 本身无状态，无需共享 session；引入外部中间件显著增加部署复杂度。

**数据持久化原则**：若未来需要持久化少量数据（如操作日志），写入容器内本地文件（`/data/`），通过 docker volume 挂载。当前 v1 无需持久化存储。

## Risks / Trade-offs

**[R1] Ceph Admin OPS API 签名兼容性** → Ceph v18.4 Admin API 使用 AWS v4 签名但 service 参数需为 `s3`；已通过 awscurl 验证可行。风险低。

**[R2] Presigned URL 跨域（CORS）** → 前端直接 PUT 到 Ceph 需要 RGW 配置 CORS 允许相应 Origin。Ceph RGW 支持 per-bucket CORS 配置，需在部署文档中说明。  
→ 缓解：提供 `s3cmd setcors` 或等效命令示例。

**[R3] 大文件上传 v1 无分片** → 单次 Presigned PUT 理论上有 5GB 上限（S3 协议限制）。v1 接受此限制，v2 可扩展为 Presigned Multipart Upload。

**[R4] JWT 泄漏风险** → 24h 有效期无 refresh 机制；内部工具场景可接受；建议部署时通过 HTTPS 反代。

**[R5] nginx 反代与 Presigned URL 的 Host 冲突** → 前端通过 nginx 反代访问 `/api`，而 Presigned URL 直连 Ceph host。两者 host 不同，浏览器不存在 cookie/session 混乱问题；但需确保 Presigned URL 中的 host 为浏览器可直接访问的 Ceph endpoint（非 backend 容器内部地址）。  
→ 缓解：后端生成 Presigned URL 时使用 `CEPH_ENDPOINT` 环境变量（浏览器可达地址），而非容器内部通信地址。

## Open Questions

- Ceph RGW 的 CORS 配置：是否需要在 README 中提供具体配置命令？（建议是）
- 是否需要中文国际化（i18n）？（默认中文 UI）
- docker-compose 是否需要内置 MinIO service 作为本地开发 mock？（可选，`STORAGE_TYPE=minio` 切换，待 MinIO provider 实现后支持）
- `STORAGE_TYPE` 环境变量是否需要在 v1 暴露，还是硬编码 `ceph`？（建议暴露，便于测试）
