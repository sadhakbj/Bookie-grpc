package bookstore

import (
	"sync"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

// InMemoryRepository stores books in a slice; kept as a documented test
// double while the production store lives elsewhere.
type InMemoryRepository struct {
	// gRPC handles concurrent requests, so every access to books is guarded.
	mu    sync.RWMutex
	books []*bookiePb.Book
}

// NewInMemoryRepository builds a repository backed by the given slice.
func NewInMemoryRepository(books []*bookiePb.Book) *InMemoryRepository {
	return &InMemoryRepository{books: books}
}

// List returns the backing slice.
func (r *InMemoryRepository) List() []*bookiePb.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.books
}

// GetByID returns the book with the given id, or false when absent.
func (r *InMemoryRepository) GetByID(id string) (*bookiePb.Book, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, book := range r.books {
		if book.Id == id {
			return book, true
		}
	}

	return nil, false
}

// Create appends the book to the backing slice.
func (r *InMemoryRepository) Create(book *bookiePb.Book) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.books = append(r.books, book)
}
