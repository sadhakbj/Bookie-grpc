package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
)

var bookClient *books.GRPCClient

func main() {
	var err error
	bookClient, err = books.NewGRPCClient()
	if err != nil {
		log.Fatalf("Failed to initialize BookClient: %v", err)
	}
	defer bookClient.Close()

	http.HandleFunc("/books", listBooksHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := bookClient.GetBooks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get books: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert the books into JSON
	booksJSON, err := json.Marshal(books)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal books to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response
	_, err = w.Write(booksJSON)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get books: %v", err), http.StatusInternalServerError)
	}
}
