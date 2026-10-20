---
status: accepted
---

# Books reads run on MySQL behind the repository seam

ADR 0002 deferred error handling until a real store needed it; sqlc now gives the books
store a typed MySQL query surface and the seed data ships as a migration. This record
freezes the MySQL-only cutover mechanics as the first persistent read path lands.

## Decision

- The server always uses the MySQL store. There is no runtime store toggle and no
  in-memory production backend; the in-memory implementation is retired. Tests mock the
  repository seam instead.
- `bookstore.Repository` evolves from context-free, error-free methods to:
  - `List(ctx) ([]*Book, error)`
  - `GetByID(ctx, id) (*Book, bool, error)`
  - `Create(ctx, *Book) error`
  The found-bool keeps ADR 0002's not-found semantics; the error carries persistence
  failures. Handlers map errors to `codes.Internal`, absent rows to `codes.NotFound`.
- The MySQL store uses only the generated sqlc queries. Canonical lowercase UUID strings
  convert to raw RFC 9562 bytes with the stdlib `uuid` package at the store boundary;
  no third-party UUID library enters the module.
- Demo data ships as goose migration `0002_seed_books.sql`: fixed valid UUIDv7 literals
  inserted as `UUID_TO_BIN(...)` (RFC byte order, no swap) with an idempotent upsert, so
  re-running never duplicates rows. Down deletes exactly those ids.
- Startup opens the pool, applies documented pool limits, pings with a short timeout,
  and fails fast with a clear log when MySQL is unreachable. The pool closes on graceful
  shutdown.

## Considered alternatives

- **Keep a memory/mysql runtime toggle (parallel bridge)** — rejected after the cutover:
  two live backends keep divergent behavior and extra config surface alive; the mock
  repository covers tests without a second implementation.
- **Auto-select MySQL when the ping succeeds, memory otherwise** — rejected: a typo or
  outage silently changes the data source instead of failing visibly.
- **Sentinel `ErrNotFound` instead of the found-bool** — rejected for continuity with
  ADR 0002; the bool already carries exactly the not-found distinction handlers need.
- **A limit parameter on `List`** — deferred: the generated query has a limit, the
  repository applies a documented default cap until a pagination ticket needs more.
- **Seed via application startup code** — rejected: seed data is schema-adjacent,
  versioned, and must be reproducible on a fresh volume through goose.

## Consequences

- Service handlers gain error paths; tests use the generated mock repository.
- The server requires `DB_*` and a reachable MySQL at startup; local runs need MySQL up
  and the migration applied (`make mysql-up && make migrate-up`).
- Docker-compose forwards `DB_*` to the gRPC server and waits for MySQL health; full
  migration bootstrap in compose remains ticket #23.
- DB `created_at`/`updated_at` are mapped but not yet exposed on the proto surface;
  pagination and created timestamps stay out of this cutover.
