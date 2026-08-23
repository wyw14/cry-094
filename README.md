# ScriptScope

ScriptScope is an offline-first operations script dependency analysis and pre-execution checking platform. It accepts immutable Shell and PowerShell artifacts, statically extracts external commands such as `jq`, `yq`, and `ansible`, correlates input/output contracts with a target inventory, and produces a dependency graph with blocking evidence. Uploaded scripts are never executed or sourced.

## Architecture

The Go service is divided by business responsibility instead of framework layers alone:

- `internal/domain`: team/RBAC, immutable script versions, semantic constraints, dependency graph, analysis issues, host inventory, precheck state machine, identity, and audit events.
- `internal/application`: upload, static analysis, team, authentication, cache, and precheck use cases. Interfaces are defined near these callers.
- `internal/parser`: bounded Shell and PowerShell static parsing. A conclusion is bound to both the artifact digest and parser build.
- `internal/repository/postgres`: pgx repositories, serializable transaction runner, cursor pagination, optimistic updates, and append-only audit persistence.
- `internal/adapter/local`: deterministic host probe and local HMAC signer used by offline demos.
- `internal/platform`: replaceable clock/ID, object storage, idempotency, and cancellable outbox worker.
- `internal/transport/http`: Gin handlers under `/api/v1`, stable error envelopes, request IDs, timeout, recovery, CORS, security headers, rate limiting, metrics, and health endpoints.
- `migrations` and `scripts/seed.sql`: versioned PostgreSQL schema and repeatable demo data.
- `api/openapi`: OpenAPI 3.0 contract.
- `web`: Vue 3, TypeScript, Vite, Pinia, Element Plus operations workspace.
- `tests/integration`: PostgreSQL transaction tests enabled with `TEST_DATABASE_URL`.

The default binary uses deterministic in-memory adapters so the service can run without a network or cloud dependency. PostgreSQL repositories implement the same application ports and are available for deployed wiring.

## Core Workflows

1. An authenticated analyst creates or joins a team and gets server-side RBAC permissions.
2. The analyst creates a script library and uploads a `.sh` or `.ps1` version. The upload boundary checks extension, MIME, size, encoding, SHA-256 digest, safe storage key, team ownership, and immutable version identity.
3. The analysis use case retrieves immutable bytes, verifies their digest, selects a bounded parser, and extracts commands, environment variables, warnings, and input/output contracts without running the content.
4. The graph builder correlates producers and consumers, evaluates the target host inventory, detects missing dependencies, unsafe commands, cycles, and incompatible versions, and retains a human-readable evidence chain.
5. A graph with a cycle or an unsatisfied blocking requirement cannot generate a precheck plan. A ready graph produces read-only inspection commands only; it never installs software.
6. The author submits the plan, a different authorized reviewer approves or rejects it with a reason, and only an approved digest can receive a signed version.
7. Every write produces an immutable audit event with actor, source, before/after state, reason, and UTC time. Compensating events are appended rather than rewriting history.

Precheck state machine:

```text
draft -> in_review -> approved -> signed
                  \-> rejected -> in_review
```

## Local Setup

Prerequisites: Go 1.24+, Node.js 20+, npm, and optionally PostgreSQL 17 and Docker Desktop.

```bash
cp .env.example .env
go mod download
go run ./cmd/server
```

In another terminal:

```bash
cd web
npm ci
npm run dev
```

Open `http://localhost:5173`. The deterministic demo login is `admin@example.test` / `scriptscope-demo`. It is for local development only; configure secrets and provision users through PostgreSQL in a deployment.

To enable PostgreSQL schema and seed data:

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/000001_initial.up.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f scripts/seed.sql
```

All stored timestamps are UTC. The UI renders them in the configured display timezone (`Asia/Shanghai` in the demo).

## Containers

```bash
docker compose up --build
```

Compose starts the application and PostgreSQL. The application image is built for the requested `TARGETOS/TARGETARCH`, then runs as an unprivileged user with a read-only root filesystem. For a multi-platform registry build:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t scriptscope:local .
```

## API Examples

Acquire a short-lived access token:

```bash
curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.test","password":"scriptscope-demo"}'
```

Create a team with the returned token:

```bash
curl -s http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Platform Operations"}'
```

Error responses have a stable shape:

```json
{"error":{"code":"VALIDATION_ERROR","message":"request fields are invalid","fields":[{"field":"Name","message":"value failed min validation"}],"request_id":"..."}}
```

See `api/openapi/openapi.yaml` for routes and parameters. Downloads always go through authorization and never return the backing storage key.

## Verification

```bash
gofmt -l cmd internal tests
go test ./...
go test -race ./...
go vet ./...
cd web && npm test
cd web && npm run typecheck
cd web && npm run build
```

The PostgreSQL integration suite skips when `TEST_DATABASE_URL` is absent. With a migrated test database it covers a forced serializable transaction rollback. Domain tests cover RBAC, last-admin protection, semantic constraints, graph cycles and evidence, precheck transitions, parser cancellation/size limits, UTF-16 upload normalization, HTTP error/request ID behavior, retry cancellation, cache concurrency, and security headers.

Verified on 2026-08-24 with Go 1.26.2 (module target Go 1.24), npm tests, Vue type checking, and the Vite production build. Docker verification is run separately by the release pipeline.

## Security And Limits

- Uploaded scripts are treated as data and never passed to a shell, `source`, PowerShell, Ansible, or an interpreter.
- Parsing is limited to 1 MiB, scanner depth, request deadline, and context cancellation. The current parser intentionally extracts a conservative command subset rather than implementing a complete shell grammar.
- Access tokens are short-lived. Refresh tokens are stored by digest and revoked on use. Passwords use bcrypt.
- Credentials, passwords, bearer/refresh tokens, raw signatures, and attachment content are excluded from structured logs.
- Local object storage is an offline adapter. Production deployments should provide encrypted durable storage through the same interface.
- The local host probe is deterministic and receives capability data; it does not connect to real hosts or persist credentials.
- The frontend currently imports Element Plus as a complete bundle, so Vite reports a non-blocking chunk-size warning. Functional tests, strict type checking, and production generation still pass.
