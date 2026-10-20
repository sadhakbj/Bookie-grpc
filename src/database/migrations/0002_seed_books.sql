-- +goose Up
-- Fixed UUIDv7 ids keep demo data deterministic across runs; UUID_TO_BIN stores
-- them in RFC byte order (no swap flag). The upsert keeps re-runs idempotent.
-- VALUES() is deprecated upstream but sqlc's migration parser rejects the newer
-- row-alias form.
INSERT INTO books (id, title, description, author, price)
VALUES
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000001'),
     'Harry Potter and the Philosopher''s Stone',
     'The boy who lived discovers he is a wizard and begins his first year at Hogwarts.',
     'J.K. Rowling', 150),
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000002'),
     '1984',
     'Winston Smith rebels against a totalitarian surveillance state ruled by Big Brother.',
     'George Orwell', 130),
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000003'),
     'To Kill a Mockingbird',
     'Atticus Finch defends an innocent man in Depression-era Alabama, told through his daughter Scout''s eyes.',
     'Harper Lee', 180),
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000004'),
     'The Pragmatic Programmer',
     'Timeless advice on crafting maintainable software and owning your career.',
     'Andrew Hunt & David Thomas', 450),
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000005'),
     'A Brief History of Time',
     'Hawking explains space, time, black holes, and the origin of the universe for general readers.',
     'Stephen Hawking', 220),
    (UUID_TO_BIN('0199e26a-2951-7000-8000-000000000006'),
     'Clean Code',
     'A handbook of agile software craftsmanship with concrete refactoring practices.',
     'Robert C. Martin', 380)
ON DUPLICATE KEY UPDATE
    title = VALUES(title),
    description = VALUES(description),
    author = VALUES(author),
    price = VALUES(price);

-- +goose Down
DELETE FROM books
WHERE id IN (
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000001'),
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000002'),
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000003'),
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000004'),
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000005'),
    UUID_TO_BIN('0199e26a-2951-7000-8000-000000000006')
);
