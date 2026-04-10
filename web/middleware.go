package web

import (
	"context"
	"net/http"
)

type contextKey string

const ctxKeyUsername contextKey = "username"

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		sess, ok := s.sessions.get(cookie.Value)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUsername, sess.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func usernameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyUsername).(string)
	return v
}
