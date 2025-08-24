package main

import (
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response with the given status code and data
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// JSONSuccess sends a successful JSON response
func JSONSuccess(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// JSONError sends an error JSON response
func JSONError(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// JSONErrorWithDetails sends an error JSON response with additional details
func JSONErrorWithDetails(w http.ResponseWriter, statusCode int, message string, details interface{}) {
	JSONResponse(w, statusCode, map[string]interface{}{
		"success": false,
		"message": message,
		"details": details,
	})
}
