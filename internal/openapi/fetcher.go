package openapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var fetchClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		if token := via[0].Header.Get("Authorization"); token != "" {
			req.Header.Set("Authorization", token)
		}
		return nil
	},
}

func Fetch(rawURL, token string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := fetchClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("spec endpoint returned HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	ct := detectContentType(resp.Header.Get("Content-Type"), rawURL)
	return data, ct, nil
}

func detectContentType(contentType, rawURL string) string {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "json") {
		return "json"
	}
	if strings.Contains(ct, "yaml") || strings.Contains(ct, "yml") {
		return "yaml"
	}

	lower := strings.ToLower(rawURL)
	if strings.HasSuffix(lower, ".json") {
		return "json"
	}
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
		return "yaml"
	}

	if len(contentType) > 0 && strings.Contains(ct, "text/plain") {
		return "yaml"
	}

	return "json"
}
