// Entrypoint for the client
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sadhakbj/bookie-grpc/src/internal/client/controllers"
	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

var bookClient *books.GRPCClient

func main() {
	logger := utils.InitializeLogger("bookie-client", true)

	var err error
	bookClient, err = books.NewGRPCClient()
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

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info("Server started on :8080")
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
