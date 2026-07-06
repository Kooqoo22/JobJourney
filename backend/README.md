# JobJourney — Backend

Job-application tracker API. Go 1.25 + Gin + PostgreSQL, Clean Architecture (see `../CLAUDE.md`).

## Layout

```
backend/
├── cmd/api/              # Composition root (main.go)
├── config/               # Env-driven typed config
├── internal/
│   ├── database/         # Connection pool + transaction manager
│   ├── middleware/       # RequestID, Recovery, Logger, CORS, Auth, RBAC, ErrorHandler
│   └── router/           # Route registration + health check
├── pkg/
│   ├── token/            # JWT access/refresh manager
│   └── utils/            # Response envelope, errors, cursor pagination
├── migrations/init.sql   # Schema (applied by docker-compose on first boot)
├── bruno/                # API collection (mirrors every endpoint)
├── docs/                 # API specification
├── Dockerfile
└── docker-compose.yml
```

Dependencies point inward: `handler → usecase → repository`. Domain packages under `internal/<domain>/` are added per feature.

## Run with Docker

```sh
cp .env.example .env
docker compose up -d --build
curl http://localhost:8080/health
```

Re-apply the schema from scratch:

```sh
docker compose down -v && docker compose up -d --build
```

## Run locally (Go)

Requires a reachable Postgres (e.g. `docker compose up -d db`).

```sh
cp .env.example .env
go run ./cmd/api
```

## Development

```sh
go fmt ./...
go vet ./...
go build ./...
```

## Configuration

All configuration is env-driven; see `.env.example`. `JWT_SECRET` is required.
