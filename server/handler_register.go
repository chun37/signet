package server

import (
	"encoding/json"
	"net/http"
)

// handleRegister はノード登録を処理する
// リクエスト: {"node_name": "alice", "nick_name": "アリス", "address": "10.0.0.1", "public_key": "..."}
// レスポンス: {"status": "registered", "block": {...}}
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeName  string `json:"node_name"`
		NickName  string `json:"nick_name"`
		Address   string `json:"address"`
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	block, err := s.node.RegisterNode(req.NodeName, req.NickName, req.Address, req.PublicKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to register node: "+err.Error())
		return
	}

	// 成功したらブロックをブロードキャスト
	s.node.BroadcastBlock(block)

	type response struct {
		Status string `json:"status"`
		Block  *Block `json:"block"`
	}
	writeJSON(w, http.StatusOK, response{
		Status: "registered",
		Block:  block,
	})
}
