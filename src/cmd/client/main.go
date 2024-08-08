// Entrypoint for the client
package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/sadhakbj/bookie-grpc/src/internal/client/controllers"
	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

var bookClient *books.GRPCClient

func main() {
	logger := utils.InitializeLogger("bookie-client", true)
	logger.Error("Starting client", slog.String("port", "8080"))

	slog.Info("Starting client", slog.String("port", "8080"))

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

	booksController, err := controllers.NewBookController()
	if err != nil {
		logger.Error("Failed to initialize BookController: %v", "error", err)
	}

	mux.HandleFunc("GET /books/{id}", booksController.FetchBookByID)
	mux.HandleFunc("GET /books", booksController.FetchAllBooks)
	logger.Info("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
