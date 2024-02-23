// Entrypoint for the client
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
)

var bookClient *books.GRPCClient

func main() {
	var err error
	bookClient, err = books.NewGRPCClient()
	if err != nil {
		log.Fatalf("Failed to initialize BookClient: %v", err)
	}
	defer func() {
		if err := bookClient.Close(); err != nil {
			log.Fatalf("Error while closing bookClient: %v", err)
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /books/{id}", listBookByIDHandler)
	mux.HandleFunc("GET /books", listBooksHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func jsonResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
	response := map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response to JSON: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("Failed to write response: %v\n", err)
	}
}

func grpcErrorToHTTPStatus(err error) (int, string) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return http.StatusNotFound, "Item not found"
		default:
			return http.StatusInternalServerError, "Something went wrong"
		}
	}
	return http.StatusInternalServerError, "Something went wrong"
}

func handleGRPCError(w http.ResponseWriter, err error) {
	log.Println("Error:", err)
	status, msg := grpcErrorToHTTPStatus(err)

	jsonResponse(w, status, false, msg, nil)
}

func listBookByIDHandler(w http.ResponseWriter, req *http.Request) {
	book, err := bookClient.GetByID(req.PathValue("id"))
	if err != nil {
		log.Println("Error:", err)
		handleGRPCError(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, true, "Fetched data successfully", []interface{}{book})
}

func listBooksHandler(w http.ResponseWriter, _ *http.Request) {
	books, err := bookClient.GetBooks()
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, true, "Successfully fetched books", books)
}
