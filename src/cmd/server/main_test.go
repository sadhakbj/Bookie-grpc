package main

import (
	"context"
	"testing"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

func TestBookieServiceGetByID(t *testing.T) {
	svc := &bookieService{}

	tests := []struct {
		name      string
		id        string
		wantCode  codes.Code
		wantTitle string
	}{
		{
			name:     "empty id is invalid argument",
			id:       "",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "malformed id is invalid argument",
			id:       "1234",
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "valid unknown uuid is not found",
			id:       "0199e26a-2951-7000-8000-00000000dead",
			wantCode: codes.NotFound,
		},
		{
			name:      "seeded book is returned",
			id:        seedHarryPotterID,
			wantCode:  codes.OK,
			wantTitle: "Harry Potter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.GetByID(context.Background(), &bookiePb.GetByIDRequest{Id: tt.id})

			if status.Code(err) != tt.wantCode {
				t.Errorf("error code = %d, want %d", status.Code(err), tt.wantCode)
			}
			if tt.wantCode == codes.OK && res.GetBook().GetTitle() != tt.wantTitle {
				t.Errorf("title = %q, want %q", res.GetBook().GetTitle(), tt.wantTitle)
			}
		})
	}
}

func TestBookieServiceCreateBook(t *testing.T) {
	origBooks := books
	defer func() { books = origBooks }()

	svc := &bookieService{}
	first, err := svc.CreateBook(context.Background(), &bookiePb.CreateBookRequest{
		Title:  "New Book",
		Price:  50,
		Author: "New Author",
	})
	if err != nil {
		t.Fatalf("CreateBook returned error: %v", err)
	}

	if first.GetBook().GetTitle() != "New Book" {
		t.Errorf("title = %q, want %q", first.GetBook().GetTitle(), "New Book")
	}
	if _, err := uuid.Parse(first.GetBook().GetId()); err != nil {
		t.Errorf("id %q is not a parseable uuid: %v", first.GetBook().GetId(), err)
	}
	parsed, err := uuid.Parse(first.GetBook().GetId())
	if err != nil || parsed.String() != first.GetBook().GetId() {
		t.Errorf("id = %q, want canonical lowercase uuid", first.GetBook().GetId())
	}
	if got := len(books); got != len(origBooks)+1 {
		t.Errorf("books length = %d, want %d", got, len(origBooks)+1)
	}

	second, err := svc.CreateBook(context.Background(), &bookiePb.CreateBookRequest{
		Title:  "Another Book",
		Price:  75,
		Author: "New Author",
	})
	if err != nil {
		t.Fatalf("second CreateBook returned error: %v", err)
	}
	if second.GetBook().GetId() == first.GetBook().GetId() {
		t.Errorf("consecutive CreateBook calls produced identical id %q", first.GetBook().GetId())
	}
}
