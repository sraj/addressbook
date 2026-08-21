# Folder Structure

## Current layout

```
cmd/
├── server/                      # HTTP server entrypoint (main.go, metrics.go)
└── admin/                       # admin CLI (migrate, seed, setup)

internal/
├── app/                         # composition root — samber/do container
│   └── app.go                   # builds injector, resolves handlers, RegisterRoutes
├── config/                      # env-var config loading
├── shared/                      # cross-cutting helpers (platform-owned)
│   ├── search.go                # port: SearchIndex interface
│   ├── db.go                    # NewDB (connection setup)
│   ├── middleware.go            # JWTAuth, UserIDFromCtx, TokenValidator
│   └── response.go              # SendError, PaginatedResponse, Recovery
├── auth/                        # bounded context — identity & JWT
│   ├── domain/                  # entity.go, errors.go, repository.go (port)
│   ├── application/             # service.go (business logic), dto.go, ports.go (provider-side ports)
│   ├── infrastructure/          # repository impl (SQLite/Postgres)
│   └── interfaces/              # handler.go, routes.go, ports.go (consumer ports), register.go (DI)
├── billing/                     # bounded context — Stripe plans/subscriptions
│   ├── domain/
│   ├── application/             # service.go, plans.go, usage.go, dto.go, ports.go
│   ├── infrastructure/          # sqlite repo + stripe client
│   └── interfaces/              # handler.go, routes.go, webhook.go, ports.go, register.go
├── features/                    # feature modules (grouped)
│   ├── contact/
│   │   ├── domain/
│   │   ├── application/         # service.go, dto.go, ports.go (quotaChecker)
│   │   ├── infrastructure/
│   │   └── interfaces/          # handler.go, routes.go, ports.go, register.go
│   ├── note/                    # (same layout)
│   └── bookmark/                # (same layout)
├── admin/                       # admin panel (services only, no domain layer)
│   ├── application/             # service.go, models.go, ports.go (billingEnsurer)
│   └── interfaces/              # handler.go, routes.go, ports.go, register.go
└── mailer/                      # email templates + Mailgun/SendGrid client

migrations/
├── sqlite/                      # driver-specific SQL migrations, per-context dirs
│   ├── auth/
│   ├── contact/
│   ├── note/
│   ├── bookmark/
│   └── billing/
├── postgres/                    # (same per-context layout)
└── embed.go                     # Contexts list, FS/Path/Table, MigrateUp/MigrateDown

resources/
└── logging/                     # optional observability stack (OpenObserve + Fluent Bit)
    ├── docker-compose.logging.yml
    ├── fluent-bit.conf
    └── dashboards/

web/                             # React SPA
```

## Dependency injection (samber/do)

`internal/api` builds a `do.Injector` container. Each bounded context's
`interfaces/register.go` exposes a `Provide(i do.Injector)` function that
registers its service + handler providers, so contexts wire their own deps
instead of one central struct being edited by everyone.

```go
// internal/auth/interfaces/register.go
func Provide(i do.Injector) {
    do.Provide(i, func(i do.Injector) (*application.Service, error) {
        db := do.MustInvoke[*xdb.DB](i)
        cfg := do.MustInvoke[*config.Config](i)
        return application.NewService(infrastructure.NewSQLiteRepo(db), cfg.JWTSecret), nil
    })
    do.Provide(i, func(i do.Injector) (*Handler, error) { ... })
    do.As[*application.Service, shared.TokenValidator](i) // service doubles as JWT validator
}
```

Adding a new context: create `interfaces/register.go` with `Provide`, then add
it to `do.New(...)` in `app.go` (which resolves its handler via
`do.MustInvoke`).

## File organization — one concern per file

Each context's `application/` and `interfaces/` packages are split by concern
so the dependency contract is visible at a glance and changes stay isolated:

```
application/                     interfaces/
├── service.go   # Service +     ├── handler.go   # Handler + HTTP methods
│                #   business    ├── routes.go    # RegisterRoutes
│                #   logic       ├── webhook.go   # webhook handlers (billing)
├── plans.go     # plan methods  ├── ports.go     # consumer ports the handler needs
├── usage.go     # quota/usage   │                #   (billingService, mailSender…)
├── dto.go       # request/      └── register.go  # samber/do Provide
│                #   response
└── ports.go     # provider ports the service needs (quotaChecker)
```

Rules:

- **ports.go** (in each layer) holds only interface definitions — the "ports"
  in ports-and-adapters terms. Consumer ports live in `interfaces/ports.go`;
  provider ports (interfaces a service depends on, e.g. `quotaChecker`,
  `billingEnsurer`) live in `application/ports.go`.
- **handler.go / service.go** hold only implementations — no interfaces mixed in.
- **routes.go** holds `RegisterRoutes` (HTTP verb → method mapping).
- **webhook.go** (billing only) holds Stripe webhook handlers.
- A service/handler may depend on several context ports; none of them import
  the concrete service package of another context — they use local interfaces.

## Migrations

Each bounded context owns its migration directory and bookkeeping table, so
teams never collide on version numbers or ordering:

- `migrations/sqlite/auth/000001_*.up.sql` → `schema_migrations_auth`
- `migrations/sqlite/contact/000001_*.up.sql` → `schema_migrations_contact`

`migrations.MigrateUp(db, driver)` runs every context in order
(`Contexts` list). Add a context by creating its dir + a `Context` entry.

## Architecture

**DDD vertical slices with ports & adapters inside each context.**

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/server/main.go                   │
├─────────────────────────────────────────────────────────┤
│  internal/api/api.go              ← composition root    │
├──────────────────────┬──────────────────────────────────┤
│    internal/auth/    │   internal/features/contact/     │
│  ┌────────────────┐  │ ┌────────────────────────────┐   │
│  │ interfaces/    │  │ │ interfaces/                │   │
│  │ application/   │  │ │ application/               │   │
│  │ infrastructure/│  │ │ infrastructure/            │   │
│  │ domain/        │  │ │ domain/                    │   │
│  └────────────────┘  │ └────────────────────────────┘   │
├──────────────────────┴──────────────────────────────────┤
│  internal/shared/     ← cross-cutting ports + helpers    │
└─────────────────────────────────────────────────────────┘
```

Dependencies flow one way: `interfaces → application → infrastructure → DB`, with `domain` at the bottom. Each layer knows only the layer below it through interfaces.

## Adding a new bounded context

```
├── features/payments/            # new
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   └── interfaces/register.go   # Provide(i do.Injector)
```

Steps:
1. Create the four layer dirs + `interfaces/register.go` with a `Provide(do.Injector)` that registers the service and handler.
2. Add a `migrations/{sqlite,postgres}/payments/` dir with 000001+ migrations.
3. Register the context in `internal/api/api.go`: add its `Provide` to `do.New(...)`, resolve the handler in `RestAPI`, and add it to `RegisterRoutes`.
4. Add a `Context` entry to `migrations/embed.go` (in dependency order).

`cmd/server/main.go` does not change for a new context.

## Scaling for team size

| Team size | Structure | Why |
|-----------|-----------|-----|
| 1–3 | Grouped DDD contexts (`features/contact`, `auth`, `billing`) | Simple, well-known |
| 3–10 | Independent contexts per feature | Teams own packages, fewer merge conflicts |
| 10+ | Add `shared/` platform team + dedicated infra packages | Platform team owns cross-cutting concerns |
