package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sadhakbj/bookie-grpc/src/internal/services/books"
)

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	return body
}

func TestFetchBookByID(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse func(*books.MockClient)
		wantStatus   int
		wantMsg      string
	}{
		{
			name: "fetches book by id",
			mockResponse: func(mc *books.MockClient) {
				mc.EXPECT().GetByID(gomock.Any()).Return(&books.Book{
					ID: "1234", Title: "Harry Potter", Price: 120, Author: "JK Rowling",
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Fetched data successfully",
		},
		{
			name: "not found is mapped to 404",
			mockResponse: func(mc *books.MockClient) {
				mc.EXPECT().GetByID(gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "Book with ID 999 not found"))
			},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Item not found",
		},
		{
			name: "unavailable is mapped to 500",
			mockResponse: func(mc *books.MockClient) {
				mc.EXPECT().GetByID(gomock.Any()).
					Return(nil, status.Error(codes.Unavailable, "server down"))
			},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := books.NewMockClient(gomock.NewController(t))
			tt.mockResponse(mockClient)
			controller := NewBookController(mockClient)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/books/1234", nil)
			req.SetPathValue("id", "1234")

			controller.FetchBookByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			body := decodeBody(t, rr)
			if body["message"] != tt.wantMsg {
				t.Errorf("message = %v, want %q", body["message"], tt.wantMsg)
			}
			data, ok := body["data"].([]any)
			if tt.wantStatus == http.StatusOK && (!ok || len(data) != 1) {
				t.Errorf("data = %v, want a list with one book", body["data"])
			}
		})
	}
}

func TestFetchAllBooks(t *testing.T) {
	t.Run("returns list of books", func(t *testing.T) {
		mockClient := books.NewMockClient(gomock.NewController(t))
		mockClient.EXPECT().GetBooks().Return([]*books.Book{
			{ID: "1", Title: "Harry Potter", Price: 120, Author: "JK Rowling"},
			{ID: "2", Title: "Game of life", Price: 450, Author: "Author Two"},
		}, nil)
		controller := NewBookController(mockClient)

		rr := httptest.NewRecorder()
		controller.FetchAllBooks(rr, httptest.NewRequest(http.MethodGet, "/books", nil))

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body := decodeBody(t, rr)
		if body["message"] != "Successfully fetched books" {
			t.Errorf("message = %v, want %q", body["message"], "Successfully fetched books")
		}
		data, ok := body["data"].([]any)
		if !ok || len(data) != 2 {
			t.Errorf("data = %v, want a list with two books", body["data"])
		}
	})

	t.Run("grpc failure is mapped to 500", func(t *testing.T) {
		mockClient := books.NewMockClient(gomock.NewController(t))
		mockClient.EXPECT().GetBooks().
			Return(nil, status.Error(codes.Unavailable, "server down"))
		controller := NewBookController(mockClient)

		rr := httptest.NewRecorder()
		controller.FetchAllBooks(rr, httptest.NewRequest(http.MethodGet, "/books", nil))

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})
}
