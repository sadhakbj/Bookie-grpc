package books

import (
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

const (
	harryPotterTitle = "Harry Potter"
	jkRowlingAuthor  = "JK Rowling"

	testBookID  = "0199e26a-2951-7000-8000-000000000001"
	testBook2ID = "0199e26a-2951-7000-8000-000000000002"
)

func TestGetByID(t *testing.T) {
	t.Run("maps proto response to Book", func(t *testing.T) {
		mockClient := bookiePb.NewMockBookieClient(gomock.NewController(t))
		mockClient.EXPECT().
			GetByID(gomock.Any(), &bookiePb.GetByIDRequest{Id: testBookID}).
			Return(&bookiePb.GetByIDResponse{
				Book: &bookiePb.Book{
					Id:          testBookID,
					Title:       harryPotterTitle,
					Price:       120,
					Author:      jkRowlingAuthor,
					Description: "a lovely book",
				},
			}, nil)
		client := NewGRPCClientWithClient(mockClient)

		book, err := client.GetByID(testBookID)
		if err != nil {
			t.Fatalf("GetByID returned error: %v", err)
		}
		want := Book{
			ID:          testBookID,
			Title:       harryPotterTitle,
			Price:       120,
			Author:      jkRowlingAuthor,
			Description: "a lovely book",
		}
		if *book != want {
			t.Errorf("mapped book = %+v, want %+v", *book, want)
		}
	})

	t.Run("returns grpc error unchanged", func(t *testing.T) {
		mockClient := bookiePb.NewMockBookieClient(gomock.NewController(t))
		mockClient.EXPECT().
			GetByID(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.NotFound, "Book with ID 999 not found"))
		client := NewGRPCClientWithClient(mockClient)

		book, err := client.GetByID("999")
		if book != nil {
			t.Errorf("book = %+v, want nil", book)
		}
		if status.Code(err) != codes.NotFound {
			t.Errorf("error code = %d, want %d", status.Code(err), codes.NotFound)
		}
	})
}

func TestGetBooks(t *testing.T) {
	t.Run("maps proto list to Book slice", func(t *testing.T) {
		mockClient := bookiePb.NewMockBookieClient(gomock.NewController(t))
		mockClient.EXPECT().
			ListBooks(gomock.Any(), &bookiePb.ListBookRequest{PerPage: 10}).
			Return(&bookiePb.ListBooksResponse{Books: []*bookiePb.Book{
				{Id: testBookID, Title: harryPotterTitle, Price: 120, Author: jkRowlingAuthor, Description: "lovely"},
				{Id: testBook2ID, Title: "Game of life", Price: 450, Author: "Author Two", Description: "test"},
			}}, nil)
		client := NewGRPCClientWithClient(mockClient)

		books, err := client.GetBooks()
		if err != nil {
			t.Fatalf("GetBooks returned error: %v", err)
		}
		if len(books) != 2 {
			t.Fatalf("got %d books, want 2", len(books))
		}
		if books[0].ID != testBookID || books[0].Title != harryPotterTitle || books[1].Price != 450 {
			t.Errorf("mapped books = %+v, want proto values", books)
		}
	})

	t.Run("returns grpc error unchanged", func(t *testing.T) {
		mockClient := bookiePb.NewMockBookieClient(gomock.NewController(t))
		mockClient.EXPECT().
			ListBooks(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.Unavailable, "server down"))
		client := NewGRPCClientWithClient(mockClient)

		books, err := client.GetBooks()
		if books != nil {
			t.Errorf("books = %+v, want nil", books)
		}
		if status.Code(err) != codes.Unavailable {
			t.Errorf("error code = %d, want %d", status.Code(err), codes.Unavailable)
		}
	})
}
