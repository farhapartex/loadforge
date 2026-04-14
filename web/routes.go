package web

import (
	"io/fs"
	"net/http"
)

func (s *Server) registerRoutes() {
	staticSub, _ := fs.Sub(embeddedFS, "static")
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/logout", s.handleLogout)

	s.mux.Handle("/api/logs/stream", s.requireAuth(http.HandlerFunc(s.logs.handleStream)))
	s.mux.Handle("/api/logs", s.requireAuth(http.HandlerFunc(s.handleLogClear)))
	s.mux.Handle("/api/run", s.requireAuth(http.HandlerFunc(s.handleRun)))
	s.mux.Handle("/api/status", s.requireAuth(http.HandlerFunc(s.handleStatus)))

	s.mux.Handle("/history", s.requireAuth(http.HandlerFunc(s.handleHistory)))
	s.mux.Handle("/api/history", s.requireAuth(http.HandlerFunc(s.handleHistoryDetail)))
	s.mux.Handle("/threshold-settings", s.requireAuth(http.HandlerFunc(s.handleThresholdSettings)))
	s.mux.Handle("/api/assertions", s.requireAuth(http.HandlerFunc(s.handleAssertions)))
	s.mux.Handle("/", s.requireAuth(http.HandlerFunc(s.handleHome)))
}
