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
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := utils.InitializeTelemetry(ctx, logger, "bookie-grpc"); err != nil {
		logger.Error("Failed to initialize telemetry: %v", "error", err)
	}
	defer utils.ShutdownTelemetry(context.Background(), logger)

	logger.Info("Creating a new server")
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	bookiePb.RegisterBookieServer(grpcServer, &bookieService{})

	// Register the standard gRPC health checking service so Kubernetes probes
	// (and grpc-health-probe) can verify liveness/readiness.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

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

	// Report NOT_SERVING so readiness probes fail immediately while draining
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	// Gracefully stop the server
	logger.Info("Gracefully stopping the gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped gracefully")
}
