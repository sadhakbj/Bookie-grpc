// Entrypoint for the grpc server
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/sadhakbj/bookie-grpc/src/database"
	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore"
	"github.com/sadhakbj/bookie-grpc/src/internal/config"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

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
	db, err := database.OpenMySQL(dbConfig)
	if err != nil {
		log.Fatal("Could not initialize MySQL books store: ", err)
	}
	defer func() { _ = db.Close() }()
	store := bookstore.NewMySQLRepository(db)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", appConfig.Port))
	if err != nil {
		log.Fatal("Could not listen: ", err)
	}

	logger := utils.InitializeLogger("bookie-grpc", false)

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
	server := newGRPCServer(store)

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		logger.Info("Successfully started the server on port: " + appConfig.Port)
		if e := server.Serve(listener); e != nil {
			logger.Error("Failed to serve: %v", "error", e)
		}
	}()

	// Wait for shutdown signal
	sig := <-signalChan
	logger.Info("Received signal. Initiating graceful shutdown...", "signal", sig)

	// Readiness probes fail immediately while draining.
	server.markNotServing()

	logger.Info("Gracefully stopping the gRPC server...")
	server.GracefulStop()
	logger.Info("Server stopped gracefully")
}
