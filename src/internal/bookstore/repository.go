//go:generate mockgen -source=repository.go -package=bookstore -destination=repository_mock.go

// Package bookstore is the repository seam over the books store, injected
// into the gRPC service handlers so tests can stand in for MySQL without a
// live database.
package bookstore

import (
	"context"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

// Repository is the books store surface the gRPC service handlers depend on;
// implementations deliver list, get-by-id, and create. The found flag keeps
// not-found distinct from a persistence error.
type Repository interface {
	List(ctx context.Context) ([]*bookiePb.Book, error)
	GetByID(ctx context.Context, id string) (*bookiePb.Book, bool, error)
	Create(ctx context.Context, book *bookiePb.Book) error
}
