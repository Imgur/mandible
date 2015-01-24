package main

import (
	"encoding/json"
	"net/http"
)

// Builds and writes an error response
func ErrorResponse(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)

	resp, _ := json.Marshal(
		map[string]interface{}{
			"success": false,
			"status":  status,
			"data": map[string]string{
				"error": message,
			},
		})

	w.Write(resp)
}
