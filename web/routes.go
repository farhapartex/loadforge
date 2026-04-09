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

	s.mux.Handle("/", s.requireAuth(http.HandlerFunc(s.handleHome)))
}
