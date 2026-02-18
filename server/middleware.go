package server

import (
	"encoding/json"
	"net/http"
)

// writeJSON はJSONレスポンスを書き込む
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError はエラーレスポンスを書き込む
func writeError(w http.ResponseWriter, status int, message string) {
	type errResponse struct {
		Error string `json:"error"`
	}
	writeJSON(w, status, errResponse{Error: message})
}
