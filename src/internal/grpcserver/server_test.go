package grpcserver

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore"
)

const seededBookID = "0199e26a-2951-7000-8000-000000000001"

func TestBookieServiceGetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		storeBook *bookiePb.Book
		found     bool
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
			id:        seededBookID,
			storeBook: &bookiePb.Book{Id: seededBookID, Title: "Harry Potter and the Philosopher's Stone"},
			found:     true,
			wantCode:  codes.OK,
			wantTitle: "Harry Potter and the Philosopher's Stone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := bookstore.NewMockRepository(ctrl)
			if _, err := uuid.Parse(tt.id); tt.id != "" && err == nil {
				store.EXPECT().GetByID(gomock.Any(), tt.id).Return(tt.storeBook, tt.found, nil)
			}
			svc := &bookieService{store: store}

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
	ctrl := gomock.NewController(t)
	store := bookstore.NewMockRepository(ctrl)

	var createdIDs []string
	store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, book *bookiePb.Book) error {
		createdIDs = append(createdIDs, book.GetId())
		return nil
	}).Times(2)

	svc := &bookieService{store: store}

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
	parsed, err := uuid.Parse(first.GetBook().GetId())
	if err != nil {
		t.Errorf("id %q is not a parseable uuid: %v", first.GetBook().GetId(), err)
	}
	if err == nil && parsed.String() != first.GetBook().GetId() {
		t.Errorf("id = %q, want canonical lowercase uuid", first.GetBook().GetId())
	}

	if _, err := svc.CreateBook(context.Background(), &bookiePb.CreateBookRequest{
		Title:  "Another Book",
		Price:  75,
		Author: "New Author",
	}); err != nil {
		t.Fatalf("second CreateBook returned error: %v", err)
	}

	if len(createdIDs) != 2 || createdIDs[0] == createdIDs[1] {
		t.Errorf("created ids = %v, want two distinct ids", createdIDs)
	}
}

func TestBookieServiceReturnsInternalOnStoreError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := bookstore.NewMockRepository(ctrl)
	store.EXPECT().List(gomock.Any()).Return(nil, errors.New("db down"))
	store.EXPECT().GetByID(gomock.Any(), seededBookID).Return(nil, false, errors.New("db down"))
	store.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db down"))

	svc := &bookieService{store: store}

	if _, err := svc.ListBooks(context.Background(), &bookiePb.ListBookRequest{}); status.Code(err) != codes.Internal {
		t.Errorf("ListBooks error code = %d, want %d", status.Code(err), codes.Internal)
	}
	if _, err := svc.GetByID(context.Background(), &bookiePb.GetByIDRequest{Id: seededBookID}); status.Code(err) != codes.Internal {
		t.Errorf("GetByID error code = %d, want %d", status.Code(err), codes.Internal)
	}
	if _, err := svc.CreateBook(context.Background(), &bookiePb.CreateBookRequest{Title: "Book"}); status.Code(err) != codes.Internal {
		t.Errorf("CreateBook error code = %d, want %d", status.Code(err), codes.Internal)
	}
}
