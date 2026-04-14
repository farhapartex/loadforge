package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/farhapartex/loadforge/internal/config"
)

func (s *Server) handleAssertions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.cfg.DefaultAssertions)
	case http.MethodPost:
		s.saveAssertions(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) saveAssertions(w http.ResponseWriter, r *http.Request) {
	var assertions []config.Assertion
	if err := json.NewDecoder(r.Body).Decode(&assertions); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError("invalid JSON body"))
		return
	}

	if assertions == nil {
		assertions = []config.Assertion{}
	}

	s.cfg.DefaultAssertions = assertions

	if err := s.cfg.save(s.configPath); err != nil {
		log.Printf("ERR save assertions: %v", err)
		writeJSON(w, http.StatusInternalServerError, apiError("failed to persist thresholds"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
