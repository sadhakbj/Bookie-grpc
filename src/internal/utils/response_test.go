package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestJSONResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	JSONResponse(rr, http.StatusOK, true, "Fetched data successfully", map[string]any{"id": "1"})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content type = %q, want %q", ct, "application/json")
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if body["success"] != true {
		t.Errorf("success = %v, want true", body["success"])
	}
	if body["message"] != "Fetched data successfully" {
		t.Errorf("message = %v, want %q", body["message"], "Fetched data successfully")
	}
}

func TestHandleGRPCError(t *testing.T) {
	rr := httptest.NewRecorder()
	HandleGRPCError(rr, status.Error(codes.NotFound, "Book not found"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if body["success"] != false {
		t.Errorf("success = %v, want false", body["success"])
	}
	if body["message"] != "Item not found" {
		t.Errorf("message = %v, want %q", body["message"], "Item not found")
	}
}
