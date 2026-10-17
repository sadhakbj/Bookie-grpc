// Entrypoint for the client
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sadhakbj/bookie-grpc/src/internal/client/controllers"
	"github.com/sadhakbj/bookie-grpc/src/internal/config"
	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

var bookClient *books.GRPCClient

func main() {
	logger := utils.InitializeLogger("bookie-client", true)
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	appConfig, err := config.LoadClientConfig()
	if err != nil {
		log.Fatal("Invalid client configuration: ", err)
	}

	telemetryCtx, telemetryCancel := context.WithCancel(context.Background())
	defer telemetryCancel()
	if err := utils.InitializeTelemetry(telemetryCtx, logger, "bookie-client", utils.TelemetryConfig{
		OTLPEndpoint:        appConfig.OTLPEndpoint,
		OTLPMetricsEndpoint: appConfig.OTLPMetricsEndpoint,
	}); err != nil {
		logger.Error("Failed to initialize telemetry: %v", "error", err)
	}
	defer utils.ShutdownTelemetry(context.Background(), logger)

	bookClient, err = books.NewGRPCClient(appConfig.GRPCServerAddr)
	if err != nil {
		logger.Error("Failed to initialize BookClient: %v", "error", err)
	}
	defer func() {
		if err := bookClient.Close(); err != nil {
			logger.Error("Error while closing bookClient: %v", "error", err)
		}
	}()

	mux := http.NewServeMux()

	booksController := controllers.NewBookController(bookClient)
	if err != nil {
		logger.Error("Failed to initialize BookController: %v", "error", err)
	}

	mux.HandleFunc("GET /books/{id}", booksController.FetchBookByID)
	mux.HandleFunc("GET /books", booksController.FetchAllBooks)

	// Health endpoint used by Kubernetes probes (liveness/readiness).
	// Returns 503 (not ready) when the gRPC upstream is unreachable, so
	// Kubernetes stops routing traffic to this BFF until it can serve.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		if bookClient != nil && bookClient.IsHealthy() {
			utils.JSONResponse(w, http.StatusOK, true, "ok", nil)
			return
		}
		utils.JSONResponse(w, http.StatusServiceUnavailable, false, "gRPC upstream not ready", nil)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + appConfig.HTTPPort,
		Handler: mux,
	}

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info("Server started on port: " + appConfig.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server: %v", "error", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-signalChan
	logger.Info("Received signal. Initiating graceful shutdown...", "signal", sig)

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the HTTP server
	logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error: %v", "error", err)
	} else {
		logger.Info("Server stopped gracefully")
	}
}
