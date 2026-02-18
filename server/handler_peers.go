package server

import (
	"net/http"
)

// handleGetPeers はピアノードのリストを返す
func (s *Server) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	peers := s.node.GetPeers()
	writeJSON(w, http.StatusOK, peers)
}
