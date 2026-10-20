package main

import (
	"context"
	"uuid"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore"
)

type bookieService struct {
	bookiePb.UnimplementedBookieServer
	store bookstore.Repository
}

func (s *bookieService) ListBooks(ctx context.Context, _ *bookiePb.ListBookRequest) (*bookiePb.ListBooksResponse, error) {
	books, err := s.store.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list books: %v", err)
	}

	return &bookiePb.ListBooksResponse{Books: books}, nil
}

func (s *bookieService) CreateBook(ctx context.Context, input *bookiePb.CreateBookRequest) (*bookiePb.CreateBookResponse, error) {
	newBook := &bookiePb.Book{
		Id:          uuid.NewV7().String(),
		Title:       input.Title,
		Price:       input.Price,
		Author:      input.Author,
		Description: input.Description,
	}
	if err := s.store.Create(ctx, newBook); err != nil {
		return nil, status.Errorf(codes.Internal, "create book: %v", err)
	}

	return &bookiePb.CreateBookResponse{
		Book: newBook,
	}, nil
}

func (s *bookieService) GetByID(ctx context.Context, input *bookiePb.GetByIDRequest) (*bookiePb.GetByIDResponse, error) {
	if input.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Please provide id")
	}
	if _, err := uuid.Parse(input.GetId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid id %s", input.GetId())
	}

	book, found, err := s.store.GetByID(ctx, input.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get book: %v", err)
	}
	if found {
		return &bookiePb.GetByIDResponse{Book: book}, nil
	}

	return nil, status.Errorf(codes.NotFound, "Book with ID %s not found", input.Id)
}

type grpcServer struct {
	*grpc.Server
	health *health.Server
}

func newGRPCServer(store bookstore.Repository) *grpcServer {
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	bookiePb.RegisterBookieServer(server, &bookieService{store: store})

	// The standard health service lets Kubernetes probes and grpc-health-probe
	// verify liveness/readiness.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	return &grpcServer{Server: server, health: healthServer}
}

func (s *grpcServer) markNotServing() {
	s.health.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
}
