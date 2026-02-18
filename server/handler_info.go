package server

import "net/http"

func (s *Server) handleGetInfo(w http.ResponseWriter, r *http.Request) {
	type response struct {
		NodeName string `json:"node_name"`
	}
	writeJSON(w, http.StatusOK, response{
		NodeName: s.node.GetNodeName(),
	})
}
