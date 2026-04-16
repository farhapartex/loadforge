package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/farhapartex/loadforge/internal/openapi"
)

func (s *Server) handleSpecInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		writeJSON(w, http.StatusBadRequest, apiError("url is required"))
		return
	}

	token := r.URL.Query().Get("token")

	data, _, err := openapi.Fetch(rawURL, token)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError("could not fetch spec: "+err.Error()))
		return
	}

	spec, err := openapi.Parse(data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError("could not parse spec: "+err.Error()))
		return
	}

	baseURL := spec.BaseURL
	detected := baseURL != ""

	if !detected {
		baseURL = inferBaseURL(rawURL)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"base_url":  baseURL,
		"detected":  detected,
	})
}

func inferBaseURL(specURL string) string {
	u, err := url.Parse(specURL)
	if err != nil || u.Host == "" {
		return ""
	}

	knownSuffixes := []string{
		"/swagger.json", "/swagger.yaml",
		"/openapi.json", "/openapi.yaml",
		"/api-docs.json", "/api-docs.yaml",
	}
	path := u.Path
	for _, suffix := range knownSuffixes {
		if strings.HasSuffix(path, suffix) {
			path = path[:len(path)-len(suffix)]
			break
		}
	}

	u.Path = path
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/")
}
