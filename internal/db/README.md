# internal/db

This package owns the data-access boundary. It is split in two:

- `queries/` — **hand-written SQL**, the source of truth. Edit these.
- `sqlcgen/` — **generated Go**, do not hand-edit. Regenerate with `make sqlc-generate` after touching `queries/` or the migrations.

`sqlcgen` is consumed only by repository implementations under `internal/domain/<entity>`. No use case or transport code should import it directly — keep the generated types behind the domain boundary so query shape changes stay local.

Schema lives under `migrations/` and is the input both for `migrate` and for `sqlc` (see `sqlc.yaml`).
