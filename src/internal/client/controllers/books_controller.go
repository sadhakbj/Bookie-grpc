// Package controllers provides HTTP handlers for the bookie client application.
package controllers

import (
	"net/http"

	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

// BookController handles HTTP requests related to book operations.
type BookController struct {
	bookClient *books.GRPCClient
}

// NewBookController creates a new BookController with the given gRPC client.
func NewBookController(bookClient *books.GRPCClient) *BookController {
	return &BookController{
		bookClient: bookClient,
	}
}

// FetchBookByID handles HTTP GET requests to fetch a book by its ID.
func (bc *BookController) FetchBookByID(w http.ResponseWriter, req *http.Request) {
	book, err := bc.bookClient.GetByID(req.PathValue("id"))
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Fetched data successfully", []interface{}{book})
}

// FetchAllBooks handles HTTP GET requests to fetch all books.
func (bc *BookController) FetchAllBooks(w http.ResponseWriter, _ *http.Request) {
	books, err := bc.bookClient.GetBooks()
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Successfully fetched books", books)
}
