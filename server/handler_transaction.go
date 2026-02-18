package server

import (
	"encoding/json"
	"net/http"
)

// handlePropose はトランザクション提案を処理する
// リクエスト: {"from": "alice", "to": "bob", "amount": 1000, "title": "飲み会代", "from_signature": "..."}
// レスポンス: {"status": "proposed", "message": "Transaction proposed to bob"}
func (s *Server) handlePropose(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From          string `json:"from"`
		To            string `json:"to"`
		Amount        int64  `json:"amount"`
		Title         string `json:"title"`
		FromSignature string `json:"from_signature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	data := &TransactionData{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Title:  req.Title,
	}

	if err := s.node.ProposeTransaction(data, req.FromSignature); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to propose transaction: "+err.Error())
		return
	}

	type response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	writeJSON(w, http.StatusOK, response{
		Status:  "proposed",
		Message: "Transaction proposed to " + req.To,
	})
}

// handleApprove はトランザクション承認を処理する
// リクエスト: {"id": "uuid-xxx"}
// レスポンス: {"status": "approved", "block": {...}}
func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	block, err := s.node.ApproveTransaction(req.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to approve transaction: "+err.Error())
		return
	}

	// 成功したらブロックをブロードキャスト
	s.node.BroadcastBlock(block)

	type response struct {
		Status string  `json:"status"`
		Block  *Block  `json:"block"`
	}
	writeJSON(w, http.StatusOK, response{
		Status: "approved",
		Block:  block,
	})
}

// handleReject はトランザクション拒否を処理する
// リクエスト: {"id": "uuid-xxx"}
// レスポンス: {"status": "rejected", "message": "Transaction rejected"}
func (s *Server) handleReject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if err := s.node.RejectTransaction(req.ID); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to reject transaction: "+err.Error())
		return
	}

	type response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	writeJSON(w, http.StatusOK, response{
		Status:  "rejected",
		Message: "Transaction rejected",
	})
}

// handleGetPending は承認待ちトランザクションの一覧を返す
func (s *Server) handleGetPending(w http.ResponseWriter, r *http.Request) {
	pending := s.node.ListPending()
	writeJSON(w, http.StatusOK, pending)
}
