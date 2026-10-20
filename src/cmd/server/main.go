// Entrypoint for the grpc server
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uuid"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore"
	"github.com/sadhakbj/bookie-grpc/src/internal/config"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

type bookieService struct {
	bookiePb.UnimplementedBookieServer
	store bookstore.Repository
}

func (s *bookieService) ListBooks(_ context.Context, req *bookiePb.ListBookRequest) (*bookiePb.ListBooksResponse, error) {
	fmt.Println("this is just a test")
	fmt.Println(req)
	return &bookiePb.ListBooksResponse{Books: s.store.List()}, nil
}

func (s *bookieService) CreateBook(_ context.Context, input *bookiePb.CreateBookRequest) (*bookiePb.CreateBookResponse, error) {
	newBook := &bookiePb.Book{
		Id:          uuid.NewV7().String(),
		Title:       input.Title,
		Price:       input.Price,
		Author:      input.Author,
		Description: input.Description,
	}
	s.store.Create(newBook)

	return &bookiePb.CreateBookResponse{
		Book: newBook,
	}, nil
}

func (s *bookieService) GetByID(_ context.Context, input *bookiePb.GetByIDRequest) (*bookiePb.GetByIDResponse, error) {
	fmt.Println("input is", input)
	if input.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Please provide id")
	}
	if _, err := uuid.Parse(input.GetId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid id %s", input.GetId())
	}
	if book, ok := s.store.GetByID(input.Id); ok {
		return &bookiePb.GetByIDResponse{Book: book}, nil
	}

	return nil, status.Errorf(codes.NotFound, "Book with ID %s not found", input.Id)
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal("Error loading .env file: ", err)
	}

	appConfig, err := config.LoadServerConfig()
	if err != nil {
		log.Fatal("Invalid server configuration: ", err)
	}

	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatal("Invalid database configuration: ", err)
	}
	db, err := sql.Open("mysql", dbConfig.DataSourceName())
	if err != nil {
		log.Fatal("Could not open MySQL connection: ", err)
	}
	defer func() { _ = db.Close() }()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	pingErr := db.PingContext(pingCtx)
	pingCancel()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", appConfig.Port))
	if err != nil {
		log.Fatal("Could not listen: ", err)
	}

	logger := utils.InitializeLogger("bookie-grpc", false)

	if pingErr != nil {
		logger.Warn("MySQL connection not successful", "error", pingErr)
	} else {
		logger.Info("MySQL connection successful")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, err = utils.InitializeTelemetry(ctx, logger, "bookie-grpc", utils.TelemetryConfig{
		OTLPEndpoint:        appConfig.OTLPEndpoint,
		OTLPLogsEndpoint:    appConfig.OTLPLogsEndpoint,
		OTLPMetricsEndpoint: appConfig.OTLPMetricsEndpoint,
	})
	if err != nil {
		logger.Error("Failed to initialize telemetry: %v", "error", err)
	}
	defer utils.ShutdownTelemetry(context.Background(), logger)

	logger.Info("Creating a new server")
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	bookiePb.RegisterBookieServer(grpcServer, &bookieService{
		store: bookstore.NewInMemoryRepository(bookstore.SeedBooks()),
	})

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
		logger.Info("Successfully started the server on port: " + appConfig.Port)
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
