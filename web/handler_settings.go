package web

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	s.templates.renderPage(w, "account_settings", PageData{
		Title:     "Settings",
		ActiveNav: "settings",
		Username:  usernameFromContext(r.Context()),
	})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Current string `json:"current"`
		New     string `json:"new"`
		Confirm string `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Current == "" || req.New == "" || req.Confirm == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all fields are required"})
		return
	}

	if !s.verifyPassword(req.Current) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is incorrect"})
		return
	}

	if req.New != req.Confirm {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new password and confirm password do not match"})
		return
	}

	if len(req.New) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new password must be at least 6 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.New), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	s.cfg.Password = string(hash)
	s.cfg.PasswordChanged = true

	if err := s.cfg.save(s.configPath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save config"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}
