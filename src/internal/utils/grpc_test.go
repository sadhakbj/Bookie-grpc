package utils

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const somethingWentWrong = "Something went wrong"

func TestGrpcErrorToHTTPStatus(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "NotFound maps to 404",
			err:         status.Error(codes.NotFound, "book not found"),
			wantStatus:  http.StatusNotFound,
			wantMessage: "Item not found",
		},
		{
			name:        "PermissionDenied maps to 500",
			err:         status.Error(codes.PermissionDenied, "denied"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: somethingWentWrong,
		},
		{
			name:        "plain error maps to 500",
			err:         errors.New("boom"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: somethingWentWrong,
		},
		{
			name:        "nil error maps to 500",
			err:         nil,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: somethingWentWrong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotMessage := GrpcErrorToHTTPStatus(tt.err)
			if gotStatus != tt.wantStatus {
				t.Errorf("status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("message = %q, want %q", gotMessage, tt.wantMessage)
			}
		})
	}
}
