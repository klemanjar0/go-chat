# go-chat

A Go chat service built on Fiber, Postgres (via `sqlc` + `pgx`), and Redis. Auth uses short-lived JWT access tokens with rotating refresh tokens, an access-token jti denylist, and a per-user revoke-before watermark for global logout.

## Layout

```
cmd/                             thin entrypoint, calls app.Run()
internal/
  app/                           composition root — wires deps together
  config/                        env loading, typed config
  domain/                        per-entity types, errors, repositories
    user/                        User, Repository (Postgres)
    auth/                        RefreshToken, RefreshTokenRepository, AccessTokenStore (Redis)
  usecase/                       application logic — one folder per entity, one file per use case
    auth/                        register, login, refresh, logout, logout_all, verify_token
    user/                        get_user
  transport/
    http/
      router.go                  route table
      middleware/                cross-cutting (auth, request id)
      handlers/auth/             auth HTTP handlers + DTOs
      handlers/user/             user HTTP handlers + DTOs
  db/
    queries/                     hand-written SQL (sqlc input)
    sqlcgen/                     generated Go (do not edit)
migrations/                      golang-migrate up/down pairs
pkg/                             reusable infra: postgres, redis, fiberutil, httputil, logger, auth (jwt + bcrypt + refresh), config, utilid
```

### Adding a new entity

1. `internal/domain/<entity>/` — model, errors, repository.
2. `internal/usecase/<entity>/` — `usecase.go` (struct + interfaces consumed by this package), one file per public method.
3. `internal/transport/http/handlers/<entity>/` — handler + DTOs.
4. Wire it into `internal/app/app.go` and `internal/transport/http/router.go`.
5. Register any new sentinel errors in `app.registerErrorMappings`.

The use case package depends on its own interfaces, not on concrete repos; the composition root supplies the implementations. This is what makes the layer testable without infra (see `internal/usecase/auth/refresh_test.go`).

## Quick start

```sh
cp .env.example .env.local        # then edit secrets if needed
make tools                        # install sqlc + golang-migrate
make dev-up                       # infra + migrations + sqlc gen
make run                          # start the API on :8080
```

Run `make help` for the full task list.

## Testing

```sh
make test                         # unit tests with -race
```

## Production checklist

- `JWT_SECRET` must be set (`config.Load` panics otherwise).
- Run behind TLS termination; the service speaks plain HTTP.
- The Docker image is built from `Dockerfile` — distroless, nonroot, stripped binary.
