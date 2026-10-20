//go:generate mockgen -source=repository.go -package=bookstore -destination=repository_mock.go

// Package bookstore is the repository seam over the books store, injected
// into the gRPC service handlers so a MySQL-backed implementation can be
// swapped in later without touching service code.
package bookstore

import (
	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

// Repository is the books store surface the gRPC service handlers depend on;
// implementations deliver list, get-by-id, and create.
type Repository interface {
	List() []*bookiePb.Book
	GetByID(id string) (*bookiePb.Book, bool)
	Create(book *bookiePb.Book)
}

// Seed ids are fixed UUIDv7 values so seeded data stays deterministic across
// runs; tests rely on the same literals. Generate more with uuid.NewV7() if
// needed.
const (
	SeedHarryPotterID         = "0199e26a-2951-7000-8000-000000000001"
	SeedNineteenEightyFourID  = "0199e26a-2951-7000-8000-000000000002"
	SeedMockingbirdID         = "0199e26a-2951-7000-8000-000000000003"
	SeedPragmaticProgrammerID = "0199e26a-2951-7000-8000-000000000004"
	SeedBriefHistoryID        = "0199e26a-2951-7000-8000-000000000005"
	SeedCleanCodeID           = "0199e26a-2951-7000-8000-000000000006"
)

// SeedBooks returns a fresh slice of the demo books the server boots with.
func SeedBooks() []*bookiePb.Book {
	return []*bookiePb.Book{
		{
			Id:          SeedHarryPotterID,
			Title:       "Harry Potter and the Philosopher's Stone",
			Price:       150,
			Author:      "J.K. Rowling",
			Description: "The boy who lived discovers he is a wizard and begins his first year at Hogwarts.",
		},
		{
			Id:          SeedNineteenEightyFourID,
			Title:       "1984",
			Price:       130,
			Author:      "George Orwell",
			Description: "Winston Smith rebels against a totalitarian surveillance state ruled by Big Brother.",
		},
		{
			Id:          SeedMockingbirdID,
			Title:       "To Kill a Mockingbird",
			Price:       180,
			Author:      "Harper Lee",
			Description: "Atticus Finch defends an innocent man in Depression-era Alabama, told through his daughter Scout's eyes.",
		},
		{
			Id:          SeedPragmaticProgrammerID,
			Title:       "The Pragmatic Programmer",
			Price:       450,
			Author:      "Andrew Hunt & David Thomas",
			Description: "Timeless advice on crafting maintainable software and owning your career.",
		},
		{
			Id:          SeedBriefHistoryID,
			Title:       "A Brief History of Time",
			Price:       220,
			Author:      "Stephen Hawking",
			Description: "Hawking explains space, time, black holes, and the origin of the universe for general readers.",
		},
		{
			Id:          SeedCleanCodeID,
			Title:       "Clean Code",
			Price:       380,
			Author:      "Robert C. Martin",
			Description: "A handbook of agile software craftsmanship with concrete refactoring practices.",
		},
	}
}
