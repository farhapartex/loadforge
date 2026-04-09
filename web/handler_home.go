package web

import "net/http"

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.templates.renderPage(w, "home", PageData{
		Title:     "Home",
		ActiveNav: "home",
		Username:  usernameFromContext(r.Context()),
		Data:      s.stats.snapshot(),
	})
}
