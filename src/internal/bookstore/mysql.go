package bookstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"uuid"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore/sqlcgen"
)

// defaultListLimit caps unpaginated reads until the API grows paging.
const defaultListLimit int32 = 100

var _ Repository = (*MySQLRepository)(nil)

// MySQLRepository persists books with the generated sqlc queries.
type MySQLRepository struct {
	queries *sqlcgen.Queries
}

// NewMySQLRepository builds a repository over the given database pool.
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{queries: sqlcgen.New(db)}
}

// List returns up to defaultListLimit books, oldest first.
func (r *MySQLRepository) List(ctx context.Context) ([]*bookiePb.Book, error) {
	rows, err := r.queries.ListBooks(ctx, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}

	books := make([]*bookiePb.Book, 0, len(rows))
	for _, row := range rows {
		books = append(books, bookFromRow(row))
	}

	return books, nil
}

// GetByID returns the book with the given canonical UUID string.
func (r *MySQLRepository) GetByID(ctx context.Context, id string) (*bookiePb.Book, bool, error) {
	idBytes, err := parseUUIDBytes(id)
	if err != nil {
		return nil, false, err
	}

	row, err := r.queries.GetBookByID(ctx, idBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get book by id: %w", err)
	}

	return bookFromRow(row), true, nil
}

// Create inserts the book using the caller-generated UUID.
func (r *MySQLRepository) Create(ctx context.Context, book *bookiePb.Book) error {
	idBytes, err := parseUUIDBytes(book.GetId())
	if err != nil {
		return err
	}

	err = r.queries.CreateBook(ctx, sqlcgen.CreateBookParams{
		ID:          idBytes,
		Title:       book.GetTitle(),
		Description: book.GetDescription(),
		Author:      book.GetAuthor(),
		Price:       book.GetPrice(),
	})
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}

	return nil
}

func bookFromRow(row sqlcgen.Book) *bookiePb.Book {
	return &bookiePb.Book{
		Id:          uuidString(row.ID),
		Title:       row.Title,
		Description: row.Description,
		Author:      row.Author,
		Price:       row.Price,
	}
}

func parseUUIDBytes(id string) ([]byte, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse book id %q: %w", id, err)
	}

	return parsed[:], nil
}

func uuidString(raw []byte) string {
	if len(raw) != len(uuid.UUID{}) {
		return ""
	}

	return uuid.UUID([16]byte(raw)).String()
}
