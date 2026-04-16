package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/farhapartex/loadforge/internal/specloader"
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		r.ParseForm()
	}

	sourceType := strings.TrimSpace(r.FormValue("source"))
	if sourceType == "" {
		sourceType = "openapi"
	}

	loader, err := s.loaders.get(sourceType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError(err.Error()))
		return
	}

	if s.runner.IsActive() {
		writeJSON(w, http.StatusConflict, apiError("a test is already running"))
		return
	}

	input := specloader.Input{
		URL:      strings.TrimSpace(r.FormValue("spec_url")),
		Token:    strings.TrimSpace(r.FormValue("token")),
		Profile:  strings.TrimSpace(r.FormValue("profile")),
		Duration: strings.TrimSpace(r.FormValue("duration")),
		BaseURL:  strings.TrimSpace(r.FormValue("base_url")),
	}
	if w2, err2 := strconv.Atoi(r.FormValue("workers")); err2 == nil && w2 > 0 {
		input.Workers = w2
	}

	if sourceType == "postman" {
		file, header, ferr := r.FormFile("postman_file")
		if ferr != nil {
			writeJSON(w, http.StatusBadRequest, apiError("postman_file is required for postman source"))
			return
		}
		defer file.Close()
		input.Data, err = io.ReadAll(file)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError("failed to read uploaded file"))
			return
		}
		input.Filename = header.Filename
	}

	if sourceType == "openapi" && input.URL == "" {
		writeJSON(w, http.StatusBadRequest, apiError("spec_url is required for openapi source"))
		return
	}

	log.Printf("Starting run  source=%s loader=%s", sourceType, loader.Name())

	cfg, err := loader.Load(input)
	if err != nil {
		log.Printf("Loader failed: %v", err)
		writeJSON(w, http.StatusBadRequest, apiError(err.Error()))
		return
	}

	cfg.Assertions = s.cfg.DefaultAssertions

	ref := input.URL
	if ref == "" {
		ref = input.Filename
	}

	if err := s.runner.Start(cfg, ref, s.stats.recordDone); err != nil {
		writeJSON(w, http.StatusConflict, apiError(err.Error()))
		return
	}

	s.stats.recordStart(ref)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "started",
		"source":  sourceType,
		"ref":     ref,
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
