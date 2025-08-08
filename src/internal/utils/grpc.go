// Package utils provides utility functions for HTTP and gRPC operations.
package utils

import (
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGRPCError converts gRPC errors to appropriate HTTP responses.
func HandleGRPCError(w http.ResponseWriter, err error) {
	log.Println("Error:", err)
	status, msg := GrpcErrorToHTTPStatus(err)

	JSONResponse(w, status, false, msg, nil)
}

// GrpcErrorToHTTPStatus converts gRPC error codes to HTTP status codes.
func GrpcErrorToHTTPStatus(err error) (int, string) {
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
