package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONResponse writes a standardized JSON response to the HTTP response writer.
func JSONResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
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
