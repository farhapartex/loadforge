package web

import (
	"net/http"
)

func (s *Server) handleThresholdSettings(w http.ResponseWriter, r *http.Request) {
	s.templates.renderPage(w, "settings", PageData{
		Title:     "SLA Thresholds",
		ActiveNav: "threshold-settings",
		Username:  usernameFromContext(r.Context()),
		Data:      s.cfg.DefaultAssertions,
	})
}
