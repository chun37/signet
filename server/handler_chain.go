package server

import (
	"encoding/json"
	"net/http"
)

// handleGetChain はチェーン全体をJSON配列で返す
func (s *Server) handleGetChain(w http.ResponseWriter, r *http.Request) {
	chain := s.node.GetChain()
	writeJSON(w, http.StatusOK, chain)
}

// handleReceiveBlock はブロックをJSONでデコードし、node.ReceiveBlock()で処理する
func (s *Server) handleReceiveBlock(w http.ResponseWriter, r *http.Request) {
	var block Block
	if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if err := s.node.ReceiveBlock(&block); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to receive block: "+err.Error())
		return
	}

	type response struct {
		Status string `json:"status"`
	}
	writeJSON(w, http.StatusOK, response{Status: "received"})
}
