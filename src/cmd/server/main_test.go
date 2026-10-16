package main

import (
	"context"
	"testing"

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
			name:      "existing book is returned",
			id:        "1234",
			wantCode:  codes.OK,
			wantTitle: "Harry Potter",
		},
		{
			name:     "unknown id is not found",
			id:       "9999",
			wantCode: codes.NotFound,
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
	res, err := svc.CreateBook(context.Background(), &bookiePb.CreateBookRequest{
		Title:  "New Book",
		Price:  50,
		Author: "New Author",
	})
	if err != nil {
		t.Fatalf("CreateBook returned error: %v", err)
	}

	if res.GetBook().GetTitle() != "New Book" {
		t.Errorf("title = %q, want %q", res.GetBook().GetTitle(), "New Book")
	}
	if res.GetBook().GetId() != "8910" {
		t.Errorf("id = %q, want %q", res.GetBook().GetId(), "8910")
	}
	if got := len(books); got != len(origBooks)+1 {
		t.Errorf("books length = %d, want %d", got, len(origBooks)+1)
	}
}
