# Development Guide

This guide explains how to set up, run, and troubleshoot the project in its current state.

## Stack Overview

- Go backend (`net/http`)
- SQLite database (`st.db`)
- Server-rendered templates and static assets in `ui/`
- TLS-enabled local server on port `4000`

## Prerequisites

- Go `1.20+`
- `make`
- Docker (optional, only for container workflow)

## Required Local Files

The application expects all of the following to exist before startup:

- `st.db` database file
- `.env` file with:
  - `DB_USER`
  - `DB_PASSWORD`
- TLS files:
  - `tls/cert.pem`
  - `tls/key.pem`

The easiest bootstrap command is:

```bash
make start
```

## Environment Variables

### Required

- `DB_USER`
- `DB_PASSWORD`

If either value is missing, startup fails.

### Optional but used by app

- `ADMIN_NAME`
- `ADMIN_PASSWORD`
- `ADMIN_EMAIL`

These values are read during startup admin creation logic.

## Run Locally (Recommended)

1) Bootstrap local prerequisites:

```bash
make start
```

2) Start the server:

```bash
go run ./cmd/web/
```

3) Open:

[`https://localhost:4000`](https://localhost:4000)

The server always starts with TLS and local certificates, so a browser warning is expected unless the cert is trusted locally.

## Run with Docker

Build and run:

```bash
make build
make run
```

Stop:

```bash
make stop
```

The container mapping is `4000:4000`, so access remains:

[`https://localhost:4000`](https://localhost:4000)

## Common Tasks

- Start app: `go run ./cmd/web/`
- Build container image: `make build`
- Run container: `make run`
- Stop container: `make stop`

## Troubleshooting

### Missing `.env`

Symptom: startup fails with `.env` load error.

Fix: run `make start` or create `.env` manually with `DB_USER` and `DB_PASSWORD`.

### Missing TLS files

Symptom: startup fails loading `./tls/cert.pem` or `./tls/key.pem`.

Fix: run `make start` to generate cert files.

### HTTP vs HTTPS confusion

Symptom: browser cannot connect over `http://localhost:4000`.

Fix: use `https://localhost:4000` because server uses `ListenAndServeTLS`.

### Port mismatch confusion

Current runtime uses port `4000`; prefer `4000` as the canonical app port across docs and local workflows.

## Related Docs

- Quickstart: [`../README.md`](../README.md)
- Audit and modernization backlog: [`./audit-and-refresh.md`](./audit-and-refresh.md)
- Architecture and task map: [`./architecture.md`](./architecture.md)
