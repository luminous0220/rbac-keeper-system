# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

RBAC Keeper System — an RBAC permission-management system. Monorepo with two independent apps:

- `server/` — Go backend (Gin, GORM/MySQL, Redis, Casbin v3, RocketMQ, zerolog). Module name `rbackeeper`.
- `admin/` — React 19 admin SPA (TypeScript, Vite 7, Ant Design 6, Zustand, TanStack Query, Tailwind). Package manager is **pnpm**.

There is no root build script; the two apps are built and run independently. Most documentation lives in `server/docs/开发计划.md` (the phased development plan, in Chinese) and `server/docs/RocketMQ迁移教学.md`. There are **no tests** anywhere in the repo.

## Commands

### Server (`server/`)

```bash
cd server
go build ./...     # compile check
go vet ./...       # static analysis
go run .           # run (needs MySQL/Redis/RocketMQ up per config.yaml)
```

- Config comes from `config.yaml` via viper (loaded with `config.LoadConfig(".")`); edit that file, not code, for ports/DSNs.
- Swagger docs under `server/docs/` are generated (`swag init`); the server serves them at `/swagger/index.html`.
- There is no `*_test.go` in the repo — don't assume `go test ./...` is meaningful. If you add tests, follow the existing package layout.
- `.vscode/launch.json` has a "Launch Package" config pointing at `server/main.go` for debugging.

### Admin (`admin/`)

```bash
cd admin
pnpm dev       # Vite dev server, port 9001 (VITE_PORT in .env.development)
pnpm build     # tsc -b && vite build
pnpm lint      # eslint .
pnpm preview   # vite preview
```

## Server architecture

### Startup pipeline (`main.go`) — order matters

1. `config.LoadConfig` → `global.CFG`
2. `initialize.RegisterLogger` (zerolog)
3. `RegisterMySQL`, `RegisterRedis`, `RegisterRocketMQ` (RocketMQ is required — fatal if absent; MinIO is commented out)
4. `RegisterMigrate` (GORM AutoMigrate of the 8 models)
5. `RegisterSeed` (built-in data, idempotent)
6. `casbinx.Init()` + `ReloadEnforcerPolicies()` (must run after seed, since policies derive from role_menus/user_roles)
7. Register operation-log MQ consumer, ensure topics exist, start consumers
8. Health check, then `router.Run` (blocking, graceful shutdown)

Global singletons live in `global/global.go`: `CFG`, `DB`, `Redis`, `RocketMQ` (producer), plus a top-level `AppCtx`/`AppCancel` used to stop background goroutines.

### Layering

Requests flow `router/ → middleware → controller → service → dao`, with shared `dto/`, `model/`, `constants/`, `response/`.

- `internal/model/` — GORM models. All embed `BaseModel` (created/updated/deleted timestamps only; **no audit-by fields by design**). JSON tags are snake_case.
- `internal/dto/` — request/response types; edit DTOs use **pointer fields** (`nil` = unchanged, zero = clear) and services build `map[string]any` for `Updates`.
- `internal/dao/` — GORM queries. Middle tables are replaced via `Association("...").Replace(...)`; association validation happens in service layer.
- `internal/controller/` — bind + validate via `response.BindJSON`/`BindQuery`, call service, return `response.OK`.
- `router/` — one `registerXxxRoutes` per module, wiring the middleware chain.

### Unified response envelope (`response/response.go`)

Every endpoint returns `Result[T]{code, msg, data}`. `code == 0` is success; business/HTTP errors use non-zero codes in the body (HTTP status stays 200). Services return `response.NewBizError(code, msg)`; controllers call `response.FailWithError`. Use `response.BindJSON`/`BindQuery` for automatic validation + Chinese error translation.

### Auth (`internal/service/auth.go`, `middleware/jwt.go`)

- Dual-token JWT (Access + Refresh), HS256. Claims hold only `UserID/Username/RoleKey/Type` (one primary role key); injected into gin context under `"claims"` and read via `utils.GetClaims(c)`.
- Captcha: `base64Captcha` + Redis `captcha:<id>` store, one-time.
- Login failure tracking in Redis (`fail:<username>` INCR, lock at 5 failures for 15 min); user-not-found and wrong-password return the same message to prevent enumeration.
- Logout blacklists the jti (`blacklist:<jti>`); password reset bumps `pwchanged:<uid>` so all tokens issued before it are rejected.
- `middleware.JWTAuth()` validates Access tokens, rejects Refresh tokens on API routes, and checks revocation.

### RBAC (Casbin) — `utils/casbinx/`

- Enforcer is `casbin.SyncedCachedEnforcer` with gorm-adapter, persisted to the `casbin_rule` table.
- Model (`rbac_model.conf`): `g(r.sub, p.sub) && keyMatch3(r.obj, p.obj) && r.act == p.act`. Perms are three-part strings from menus, e.g. `system:user:list`, checked as `Enforce(username, perm, "*")`.
- `ReloadEnforcerPolicies` rebuilds all `p` (role→perm) and `g` (user→role) rules from DB; the `super` role gets a `*,*` wildcard. `Reload` wipes the DB rules first, so revoking a role/perm takes effect immediately.
- **After any change to menus, roles, or role–menu associations, call `casbinx.ReloadEnforcerPolicies()`** (see `internal/service/menu.go` and `role.go`). Note: changing a user's roles via `service/user.go` currently does **not** trigger a reload — a known gap.
- `middleware.RBAC(obj, act)` → 401 if unauthenticated, 403 (with an async operation log) if denied.

### Data scope (row-level access) — 3-part mechanism

1. `middleware.DataScope()` — computes the **widest** scope across the user's roles (min of `data_scope` values; `super` is `DsAll`), writes 4 context keys (`dataScope`, `dataScopeDepts`, `dataScopeDept`, `dataScopeUserID`).
2. DAO functions apply `utils.GetUserDataScope(ctx)` / `utils.GetDataScope(ctx)` as a GORM scope on the entity being filtered.
3. Scope values (`internal/constants/data_scope.go`): `1` all, `2` custom (role_depts), `3` own dept, `4` dept+sub-depts, `5` self.

When adding a list endpoint that filters by department/user, you must (a) mount `middleware.DataScope()` after `JWTAuth()` and (b) apply the scope in the DAO query — both are easy to forget.

### Middleware chain for protected routes

Global: `Recovery()`, `RequestLog()`, `CORS()` in `router.Setup`. Per-module groups (see `router/user.go`/`dict.go`):

```go
ug := base.Group("/user")
ug.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
ug.GET("/list", middleware.RBAC("system:user:list", "*"), middleware.DataScope(), user.List)
ug.POST("/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:user:add", "*"), middleware.DataScope(), user.Add)
```

- `RateLimit(rate, burst)` — Redis token bucket (distributed).
- `CaptureBody()` — buffers the request body into context key `requestBody` so operation logs can log (masked) params; must precede any route whose controller calls `service.CreateLogWithMq`.
- `Idempotent(ttl)` — write-only, Redis `SETNX` keyed on `X-Request-Id`; releases after the request so retries aren't blocked forever.

### Operation logs (async via RocketMQ)

`service.CreateLogWithMq(c, title)` enqueues a payload to topic `"LOG"` (group `log_consumer_group`); `internal/consumer/log.go` consumes and writes to `operation_logs`. Sensitive JSON fields are masked via `utils.MaskSensitiveJSON`. Failures only `log.Warn` — they never block the business response.

### Seed data (`initialize/seed.go`)

Idempotent (skips if `super` role exists): 10 depts, 39 menus (M=directory/C=menu/F=button), 4 roles (`super`/`admin`/`user`/`staff`), 10 users. Default password for all seeded users: `123456aaa`. Login with `super` / `123456aaa` for full access. `super` username and role key are protected from deletion/edit.

### Config

`config.yaml` → `config/config.go` via viper. `time.Duration` fields are written as strings (`"10s"`, `"3000h"`). Dev secrets (MySQL/Redis passwords, JWT secret) are hardcoded in `config.yaml` — dev-only, per the plan's known-issues list.

## Admin architecture

- **Entry**: `src/main.tsx` → `App.tsx` (antd `ConfigProvider` with dark/light theme + `QueryClientProvider` + message holders).
- **Routing**: hash router (react-router 7) in `src/routers/index.tsx`. `AuthGuard` fetches permissions, stores user info + menu tree in the Zustand `auth` store, and routes are **dynamically generated from the menu tree** (`generateRoutes`). Routes change → `RouterProvider` is remounted via a `key` derived from route paths.
- **State**: Zustand stores in `src/stores/` (`auth`, `global`, `theme`, `calendar`). Reusable UI lives in `src/components/` (ProTable, ProForm, ProModal, ModalSelector, Chart, Card, Calendar, Banner, …) — the `pages/Demo/*` pages exercise these components.
- **API layer**: `src/libs/request.ts` is an axios wrapper that injects `Bearer` token from storage and unwraps `{code, msg, data}`. `src/apis/index.ts` defines typed endpoints.

> **Known mismatch**: the admin's `src/apis` still points at the upstream template's demo paths (`/system/api/User/login`, `/system/api/Menu/list`, …) served by `@meadmin-cn/vite-plugin-mock` in dev. The Go backend actually serves `/api/auth/login`, `/api/menu/tree`, etc. The SPA currently runs against mock data; wiring the admin to the real backend endpoints is unfinished work.

## Conventions (from `server/docs/开发计划.md`)

- Chinese comments throughout, explaining **why** (not what). Match this style.
- Wrap errors: `fmt.Errorf("...: %w", err)`.
- Logical delete via `gorm.DeletedAt`; middle-table cleanup via raw SQL; referential checks in the service layer.
- Controllers stay thin; business logic lives in services; SQL lives in DAOs.

## Known limitations / unfinished work (documented in the plan)

- **No test files exist** — the plan lists adding tests as a pending item.
- Audit fields (`CreateBy`/`UpdateBy`) intentionally absent from `BaseModel`.
- MinIO registration is commented out in `main.go`; file upload is not wired.
- IP geo-location (`utils.ParseIPLocation`) is a stub.
- Dict caching uses a delayed double-delete pattern in `internal/service/dict_cache.go`.
- The `super`/`admin`/`user`/`staff` seed users share one default password; change before any real use.
