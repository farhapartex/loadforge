package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

type PageData struct {
	Title     string
	ActiveNav string
	Username  string
	Error     string
	Data      any
}

type templateCache struct {
	pages map[string]*template.Template
	login *template.Template
}

func newTemplateCache() (*templateCache, error) {
	tmplFS, err := fs.Sub(embeddedFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("sub templates fs: %w", err)
	}

	pageNames := []string{"home", "history", "settings", "account_settings"}
	pages := make(map[string]*template.Template, len(pageNames))

	for _, name := range pageNames {
		t, err := template.New("").ParseFS(tmplFS, "layout.html", name+".html")
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", name, err)
		}
		pages[name] = t
	}

	loginTmpl, err := template.New("").ParseFS(tmplFS, "login.html")
	if err != nil {
		return nil, fmt.Errorf("parse login template: %w", err)
	}

	return &templateCache{pages: pages, login: loginTmpl}, nil
}

func (tc *templateCache) renderPage(w http.ResponseWriter, name string, data PageData) {
	t, ok := tc.pages[name]
	if !ok {
		http.Error(w, fmt.Sprintf("template %q not found", name), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (tc *templateCache) renderLogin(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tc.login.ExecuteTemplate(w, "login", data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}
