package main

import (
	"encoding/json"
	"net/http"
)

func Response(w http.ResponseWriter, data map[string]interface{}, status ...int) {
	var responseStatus int
	if len(status) > 0 {
		responseStatus = status[0]
	} else {
		responseStatus = http.StatusOK
	}

	w.WriteHeader(responseStatus)

	resp, _ := json.Marshal(
		map[string]interface{}{
			"success": responseStatus == http.StatusOK,
			"status":  responseStatus,
			"data":    data,
		})

	w.Write(resp)
}

func ErrorResponse(w http.ResponseWriter, message string, status int) {
	Response(w, map[string]interface{}{"error": message}, status)
}
