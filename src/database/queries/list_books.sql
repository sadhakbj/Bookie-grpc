-- ListBooks returns books oldest-first (id breaks same-second ties), capped by
-- the given limit.

-- name: ListBooks :many
SELECT id, title, description, author, price, created_at, updated_at
FROM books
ORDER BY created_at, id
LIMIT ?;
