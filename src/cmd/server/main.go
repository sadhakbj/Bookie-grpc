// Entrypoint for the grpc server
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

var books = []*bookiePb.Book{
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

type bookieService struct {
	bookiePb.UnimplementedBookieServer
}

func (s *bookieService) ListBooks(_ context.Context, req *bookiePb.ListBookRequest) (*bookiePb.ListBooksResponse, error) {
	fmt.Println("this is just a test")
	fmt.Println(req)
	return &bookiePb.ListBooksResponse{Books: books}, nil
}

func (s *bookieService) CreateBook(_ context.Context, input *bookiePb.CreateBookRequest) (*bookiePb.CreateBookResponse, error) {
	newBook := &bookiePb.Book{
		Id:          "8910",
		Title:       input.Title,
		Price:       input.Price,
		Author:      input.Author,
		Description: input.Description,
	}
	books = append(books, newBook)

	return &bookiePb.CreateBookResponse{
		Book: newBook,
	}, nil
}

func (s *bookieService) GetByID(_ context.Context, input *bookiePb.GetByIDRequest) (*bookiePb.GetByIDResponse, error) {
	fmt.Println("input is", input)
	if input.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Please provide id")
	}
	for _, book := range books {
		if book.Id == input.Id {
			return &bookiePb.GetByIDResponse{Book: book}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "Book with ID %s not found", input.Id)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8020"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Could not listen: ", err)
	}

	logger := utils.InitializeLogger("bookie-grpc", false)

	logger.Info("Creating a new server")
	grpcServer := grpc.NewServer()
	bookiePb.RegisterBookieServer(grpcServer, &bookieService{})

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		logger.Info("Successfully started the server on port: " + port)
		if e := grpcServer.Serve(listener); e != nil {
			logger.Error("Failed to serve: %v", "error", e)
		}
	}()

	// Wait for shutdown signal
	sig := <-signalChan
	logger.Info("Received signal. Initiating graceful shutdown...", "signal", sig)

	// Gracefully stop the server
	logger.Info("Gracefully stopping the gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped gracefully")
}
