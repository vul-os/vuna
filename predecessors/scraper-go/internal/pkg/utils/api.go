package utils

import (
	"encoding/json"
	"net/http"
)

func DataToJson(data interface{}, w http.ResponseWriter, statusCode int) {
	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code
	w.WriteHeader(statusCode)

	// Write the JSON response
	w.Write(jsonData)
}