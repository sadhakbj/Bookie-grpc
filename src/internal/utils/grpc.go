package utils

import (
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGRPCError(w http.ResponseWriter, err error) {
	log.Println("Error:", err)
	status, msg := GrpcErrorToHTTPStatus(err)

	JSONResponse(w, status, false, msg, nil)
}

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
