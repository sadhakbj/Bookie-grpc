-- CreateBook takes the caller-generated BINARY(16) id bytes; SQL never
-- generates ids.

-- name: CreateBook :exec
INSERT INTO books (id, title, description, author, price)
VALUES (?, ?, ?, ?, ?);
