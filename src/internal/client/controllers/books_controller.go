package controllers

import (
	"net/http"

	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
	"github.com/sadhakbj/bookie-grpc/src/internal/utils"
)

type BookController struct {
	bookClient *books.GRPCClient
}

func NewBookController(bookClient *books.GRPCClient) *BookController {
	return &BookController{
		bookClient: bookClient,
	}
}

func (bc *BookController) FetchBookByID(w http.ResponseWriter, req *http.Request) {
	book, err := bc.bookClient.GetByID(req.PathValue("id"))
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Fetched data successfully", []interface{}{book})
}

func (bc *BookController) FetchAllBooks(w http.ResponseWriter, _ *http.Request) {
	books, err := bc.bookClient.GetBooks()
	if err != nil {
		utils.HandleGRPCError(w, err)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Successfully fetched books", books)
}
