package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func jsonify(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.Write([]byte(
			"{\"success\": false, \"message\": \"failed to encode json response\"}",
		))
		w.WriteHeader(500)
		return
	}
}

func jsonifyErrf(w http.ResponseWriter, statusCode int, format string, data ...interface{}) {
	msg := fmt.Sprintf(format, data...)
	success := false
	if statusCode >= 200 && statusCode <= 400 {
		success = true
	} else {
		log.Printf("failed to respond: %v", msg)
	}

	jsonify(w, statusCode, struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: success,
		Message: msg,
	})
}
