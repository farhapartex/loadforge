package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/farhapartex/loadforge/internal/config"
)

func buildRequest(baseURL string, step config.Step) (*http.Request, error) {
	fullURL, err := resolveURL(baseURL, step.URL)

	if err != nil {
		return nil, fmt.Errorf("Invalid url %q: %w", step.URL, err)
	}

	bodyReader, contentType, err := buildBody(step.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to build request body: %w", err)
	}

	req, err := http.NewRequest(step.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	if contentType != "" {
		if _, exists := step.Headers["Content-Type"]; !exists {
			req.Header.Set("Content-Type", contentType)
		}
	}

	for key, value := range step.Headers {
		req.Header.Set(key, value)
	}

	if err := applyAuth(req, step.Auth); err != nil {
		return nil, fmt.Errorf("Failed to apply auth: %w", err)
	}

	return req, nil
}

func resolveURL(baseURL string, stepURL string) (string, error) {
	if strings.HasPrefix(stepURL, "http://") || strings.HasPrefix(stepURL, "https://") {
		return stepURL, nil
	}

	if baseURL == "" {
		return "", fmt.Errorf("Step URL %q is relative but no base url is set in config", stepURL)
	}

	if _, err := url.Parse(baseURL); err != nil {
		return "", fmt.Errorf("Invalid base url %q: %w", baseURL, err)
	}

	base := strings.TrimRight(baseURL, "/")
	path := "/" + strings.TrimLeft(stepURL, "/")

	return base + path, nil
}

func buildBody(body *config.Body) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
	}

	if body.Raw != "" {
		return bytes.NewReader([]byte(body.Raw)), "text/plain", nil
	}

	if body.JSON != nil {
		data, err := json.Marshal(body.JSON)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal json body: %w", err)
		}

		return bytes.NewReader(data), "application/json", nil
	}

	if len(body.Form) > 0 {
		form := url.Values{}
		for key, value := range body.Form {
			form.Set(key, value)
		}

		encoded := form.Encode()

		return bytes.NewReader([]byte(encoded)), "application/x-www-form-urlencoded", nil
	}

	return nil, "", nil
}

func applyAuth(req *http.Request, auth *config.Auth) error {
	if auth == nil {
		return nil
	}

	if auth.Basic != nil {
		req.SetBasicAuth(auth.Basic.Username, auth.Basic.Password)
		return nil
	}

	if auth.Bearer != "" {
		req.Header.Set("Authorization", "Bearer "+auth.Bearer)
		return nil
	}

	if auth.Header != nil {
		if auth.Header.Key == "" {
			return fmt.Errorf("auth.header.key cannot be empty")
		}
		req.Header.Set(auth.Header.Key, auth.Header.Value)
	}

	return nil
}
