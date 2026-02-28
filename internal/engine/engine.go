package engine

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
)

const MaxBodyRead = 1 * 1024 * 1024 // 1 MB

// Result holds the outcome of a single HTTP request execuation
type Result struct {
	StepName   string
	Method     string
	URL        string
	StatusCode int
	Duration   time.Duration
	BytesRead  int64
	Error      error
	Body       []byte
	Headers    http.Header
}

func (r *Result) IsSuccess() bool {
	return r.Error == nil
}

type Engine struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// ExecuteStep runs a single Step from a scenario and returns the Result
func (e *Engine) ExecuteStep(step config.Step) *Result {
	result := &Result{
		StepName: step.Name,
		Method:   step.Method,
	}

	client := buildClient(step.Options)

	req, err := buildRequest(e.cfg.BaseURL, step)
	if err != nil {
		result.Error = fmt.Errorf("failed to build request: %w", err)
		return result
	}

	result.URL = req.URL.String()

	start := time.Now()
	resp, err := client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		return result
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Headers = req.Header

	limitedReader := io.LimitReader(resp.Body, MaxBodyRead)
	body, err := io.ReadAll(limitedReader)

	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		return result
	}

	result.Body = body
	result.BytesRead = int64(len(body))

	return result
}
