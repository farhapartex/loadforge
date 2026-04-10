package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleRunStart(w, r)
	case http.MethodDelete:
		s.handleRunStop(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRunStart(w http.ResponseWriter, r *http.Request) {
	specURL  := strings.TrimSpace(r.FormValue("spec_url"))
	token    := strings.TrimSpace(r.FormValue("token"))
	profile  := strings.TrimSpace(r.FormValue("profile"))
	duration := strings.TrimSpace(r.FormValue("duration"))
	workers  := 0
	if w2, err := strconv.Atoi(r.FormValue("workers")); err == nil && w2 > 0 {
		workers = w2
	}

	if specURL == "" {
		writeJSON(w, http.StatusBadRequest, apiError("spec_url is required"))
		return
	}

	if s.runner.IsActive() {
		writeJSON(w, http.StatusConflict, apiError("a test is already running"))
		return
	}

	log.Printf("Fetching spec url=%s", specURL)

	data, _, err := s.openapi.Fetch(specURL, token)
	if err != nil {
		log.Printf("Fetch failed: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError("failed to fetch spec: "+err.Error()))
		return
	}

	spec, err := s.openapi.Parse(data)
	if err != nil {
		log.Printf("Parse failed: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError("failed to parse spec: "+err.Error()))
		return
	}

	ops := s.openapi.Extract(spec)
	if len(ops) == 0 {
		writeJSON(w, http.StatusBadRequest, apiError("no operations found in spec"))
		return
	}

	cfg, err := s.openapi.Generate(ops, spec.BaseURL, token)
	if err != nil {
		log.Printf("Generate failed: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError("failed to generate config: "+err.Error()))
		return
	}

	if workers > 0 {
		cfg.Load.Workers = workers
	}
	if duration != "" {
		cfg.Load.Duration = duration
	}
	if profile != "" {
		cfg.Load.Profile = profile
	}

	if err := s.runner.Start(cfg, specURL, s.stats.recordDone); err != nil {
		writeJSON(w, http.StatusConflict, apiError(err.Error()))
		return
	}

	s.stats.recordStart(specURL)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "started",
		"spec":    specURL,
		"workers": cfg.Load.Workers,
		"profile": cfg.Load.Profile,
	})
}

func (s *Server) handleRunStop(w http.ResponseWriter, r *http.Request) {
	if !s.runner.IsActive() {
		writeJSON(w, http.StatusConflict, apiError("no test is currently running"))
		return
	}

	s.runner.Stop()
	log.Printf("Run stop requested by %s", usernameFromContext(r.Context()))

	writeJSON(w, http.StatusOK, map[string]string{"status": "stopping"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func apiError(msg string) map[string]string {
	return map[string]string{"error": msg}
}
