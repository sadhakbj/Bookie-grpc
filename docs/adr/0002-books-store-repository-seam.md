---
status: accepted
---

# The books store hides behind a repository seam

The gRPC service handlers owned and mutated the books slice directly, so swapping storage
meant rewriting service code. We decided to carve a small `bookstore.Repository` interface
(list, get-by-id, create — exactly the operations the handlers need, nothing speculative)
and inject the store into `bookieService` via a constructor. The slice-backed
`InMemoryRepository` is one implementation of that seam, retained as a test double; the
seed fixture it boots with lives with the implementation, not in the application entrypoint.
A MySQL implementation selected by config lands in the data-layer wave without touching
service code — "make the change easy, then make the easy change".

## Considered alternatives

- **Keep the global slice and add MySQL beside it** — rejected: handlers stay entangled
  with storage mechanics and every cutover step rewrites service code.
- **Error-returning interface methods now** (`List() (…, error)`, `ErrNotFound` sentinel) —
  rejected for this prefactor: handlers cannot fail today, and inventing error paths is a
  behavior change; the signatures evolve when the MySQL ticket actually needs them.
- **Sentinel-error probing in `GetByID`** — rejected in favor of a found-bool so the
  service keeps its exact `codes.NotFound` behavior with zero new semantics.

## Consequences

- `main.go` no longer owns book state; the in-memory seed data moved into
  `src/internal/bookstore` with the implementation and is due for retirement from
  non-test paths once MySQL is live.
- The in-memory store guards its slice with an `RWMutex`: gRPC serves handlers on
  concurrent goroutines.
- The seam is mockable via mockgen (`go:generate` on the interface), ready for the
  MySQL store's tests.
