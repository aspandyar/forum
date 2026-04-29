# Forum

Simple Go-based forum application with server-rendered HTML, SQLite storage, and role-based moderation features.

## Features

- User signup/login/logout
- Forum post create/edit/delete
- Comments and likes/dislikes for posts and comments
- Category/tag filtering
- Role-based moderation flows
- Google and GitHub OAuth login paths

## Quickstart

### Prerequisites

- Go 1.20+
- Docker (optional, for container workflow)
- `make`

### 1) Bootstrap local prerequisites

This project expects three local artifacts:

- `st.db`
- `.env` (with DB credentials)
- TLS certificates under `tls/`

Create all of them with:

```bash
make start
```

### 2) Run the app

Run directly with Go:

```bash
go run ./cmd/web/
```

Open:

`[https://localhost:4000](https://localhost:4000)`

The server uses TLS, so your browser may show a certificate warning for local development.

## Docker Workflow

```bash
make build
make run
make stop
```

`make run` maps `4001:4000`, so the app remains available at:

`[https://localhost:4001](https://localhost:4001)`

The Docker workflow now uses `docker-compose.yml`, mounting:

- `.env` into container runtime environment
- `st.db` at `/app/st.db`
- `tls/` at `/app/tls` (read-only)

## Testing and Coverage

Run all tests:

```bash
make test
```

Generate coverage profile and summary:

```bash
make test-cover
```

Enforce minimum coverage threshold (default `95%`):

```bash
make test-cover-enforce
```

Override threshold when needed:

```bash
make test-cover-enforce COVERAGE_THRESHOLD=80
```

## CI/CD

- `CI` workflow runs on pull requests and pushes to `main`/`feature/*`, executes `make test-cover-enforce`, and uploads `coverage.out`.
- `CD` workflow runs on pushes to `main` and `v*` tags, then builds and pushes the Docker image to `ghcr.io/<owner>/<repo>`.

## Documentation

- Development setup, env vars, TLS details, troubleshooting: `[docs/development.md](docs/development.md)`
- Refresh audit and modernization backlog: `[docs/audit-and-refresh.md](docs/audit-and-refresh.md)`
- Architecture and important task map: `[docs/architecture.md](docs/architecture.md)`
- HTTP API (OpenAPI): `[docs/openapi.yaml](docs/openapi.yaml)`. With the server running, browse **[Swagger UI](https://localhost:4000/swagger/)** (same TLS note as the app).

## Authors

- `@aspandyar`
