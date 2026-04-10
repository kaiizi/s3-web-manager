## ADDED Requirements

### Requirement: Separate Dockerfiles for frontend and backend
The project SHALL provide independent Dockerfiles for frontend and backend. The `backend/Dockerfile` SHALL be a multi-stage build producing a minimal Go binary on alpine. The `frontend/Dockerfile` SHALL be a multi-stage build that runs `npm run build` then serves the dist via nginx:alpine with a custom `nginx.conf` that includes SPA fallback and `/api` reverse proxy to the backend service.

#### Scenario: Backend Docker build succeeds
- **WHEN** user runs `docker build -t s3-web-manager-backend ./backend`
- **THEN** build completes without error and produces a runnable alpine image with only the Go binary

#### Scenario: Frontend Docker build succeeds
- **WHEN** user runs `docker build -t s3-web-manager-frontend ./frontend`
- **THEN** build completes without error and produces an nginx:alpine image serving the Vue SPA

#### Scenario: nginx routes API to backend
- **WHEN** browser sends a request to `/api/...` on the frontend container
- **THEN** nginx reverse-proxies the request to `backend:8080`

#### Scenario: SPA fallback
- **WHEN** browser requests a non-asset path (e.g., `/buckets/my-bucket`)
- **THEN** nginx serves `index.html` so Vue Router handles the route

### Requirement: docker-compose configuration
The project SHALL provide a `docker-compose.yml` that defines `backend` and `frontend` as separate services. Both services SHALL read configuration from a shared `.env` file via `env_file`. The frontend service SHALL depend on the backend service.

#### Scenario: Start with docker-compose
- **WHEN** user runs `docker-compose up -d` with a valid `.env` file
- **THEN** both services start; the application is accessible via the frontend container's exposed port (default 80)

### Requirement: Environment variable configuration
The application SHALL read all configuration from environment variables. A `.env.example` file SHALL document all required and optional variables.

Required variables:
- `APP_USERNAME`: login username
- `APP_PASSWORD`: login password
- `JWT_SECRET`: JWT signing secret (minimum 32 characters)
- `STORAGE_TYPE`: storage provider type (`ceph`; future: `minio`, `s3`). Default: `ceph`
- `CEPH_ENDPOINT`: Ceph RGW endpoint URL (e.g., `http://172.16.31.61:8000`)
- `CEPH_ACCESS_KEY`: admin access key
- `CEPH_SECRET_KEY`: admin secret key
- `CEPH_REGION`: S3 region (default: `default`)
- `APP_PORT`: backend listen port (default: `8080`)

#### Scenario: Missing required variable
- **WHEN** application starts without a required environment variable
- **THEN** application logs a fatal error and exits with non-zero code

#### Scenario: Default port
- **WHEN** `APP_PORT` is not set
- **THEN** server listens on port `8080`

### Requirement: README documentation
The project SHALL provide a `README.md` documenting: project overview, prerequisites, quick start (docker-compose), environment variable reference, Ceph CORS configuration instructions, and local development setup.

#### Scenario: CORS configuration guidance
- **WHEN** user follows README instructions for Ceph CORS setup
- **THEN** browser can directly PUT/GET objects via Presigned URLs without CORS errors

### Requirement: No external middleware dependencies
The system SHALL NOT depend on any external middleware services (Redis, message queues, external session stores, etc.). All runtime state SHALL be held in memory. If persistent local storage is needed in future versions, it SHALL use local files mounted via Docker volume.

#### Scenario: System starts with only Docker and .env
- **WHEN** user runs `docker-compose up -d` with only the backend and frontend services defined
- **THEN** the system is fully functional without any additional infrastructure

### Requirement: S3 provider abstraction
The backend SHALL define a `StorageProvider` interface that abstracts all storage operations. The initial implementation SHALL be `CephProvider`. Provider selection SHALL be driven by the `STORAGE_TYPE` environment variable to enable future MinIO or AWS S3 support without modifying API handlers.

#### Scenario: CephProvider selected by default
- **WHEN** `STORAGE_TYPE` is `ceph` or not set
- **THEN** backend initializes and uses `CephProvider`

#### Scenario: Unknown storage type
- **WHEN** `STORAGE_TYPE` is set to an unsupported value
- **THEN** application logs a fatal error and exits
