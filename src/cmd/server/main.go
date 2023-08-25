package main

import (
	"context"
	"log"
	"net"
	"slices"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"google.golang.org/grpc"
)

type bookieService struct {
	bookiePb.UnimplementedBookieServer
}

func (s *bookieService) ListBooks(context.Context, *bookiePb.ListBookRequest) (*bookiePb.ListBooksResponse, error) {
	books := []*bookiePb.Book{
		{
			Id:          "1234",
			Title:       "Harry Potter",
			Price:       120,
			Author:      "JK Rowling",
			Description: "a lovely book",
		},
		{
			Id:          "4567",
			Title:       "Game of life",
			Price:       450,
			Author:      "Author Two",
			Description: "This is a test",
		},
	}

	books = slices.Insert(books, 2, &bookiePb.Book{
		Id:          "8910",
		Title:       "Game of Thrones",
		Price:       120,
		Author:      "Author Three",
		Description: "This is a another test",
	})

	return &bookiePb.ListBooksResponse{Books: books}, nil
}

func (s *bookieService) CreateBook(context.Context, *bookiePb.CreateBookRequest) (*bookiePb.CreateBookResponse, error) {
	return &bookiePb.CreateBookResponse{}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8020")
	if err != nil {
		log.Fatal("Could not listen: ", err)
	}

	log.Print("Creating a new server")
	grpcServer := grpc.NewServer()
	bookiePb.RegisterBookieServer(grpcServer, &bookieService{})

	log.Printf("Successfully started the server")

	if e := grpcServer.Serve(listener); e != nil {
		panic(e)
	}

}
