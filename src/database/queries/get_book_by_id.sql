-- GetBookByID matches the BINARY(16) primary key, passed as []byte holding the
-- raw UUID bytes.

-- name: GetBookByID :one
SELECT id, title, description, author, price, created_at, updated_at
FROM books
WHERE id = ?;
