package web

import (
	"net/http"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.templates.renderPage(w, "home", PageData{
		Title:     "Home",
		ActiveNav: "home",
		Username:  usernameFromContext(r.Context()),
		Data:      s.liveStats(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.liveStats())
}

func (s *Server) liveStats() RunStatsSnapshot {
	records := s.runner.Results().All()

	snap := s.stats.snapshot()
	snap.TotalRuns = len(records)

	if s.runner.IsActive() {
		snap.ActiveTests = 1
	} else {
		snap.ActiveTests = 0
	}

	if len(records) > 0 {
		latest := records[0]
		snap.LastStatus = string(latest.Status)
		snap.LastRunAt = latest.StartedAt.Format("Jan 2, 15:04")
		snap.LastConfig = latest.SpecURL
	}

	return snap
}
