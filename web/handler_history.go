package web

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/farhapartex/loadforge/internal/loader"
	"github.com/farhapartex/loadforge/internal/runner"
)

type historyRow struct {
	ID        string
	SpecURL   string
	Profile   string
	Workers   int
	Duration  string
	StartedAt string
	Elapsed   string
	Status    string
	Requests  int64
	Successes int64
	Failures  int64
	ErrorRate string
	RPS       string
}

type runDetail struct {
	ID               string
	SpecURL          string
	Profile          string
	Workers          int
	Duration         string
	StartedAt        string
	EndedAt          string
	Elapsed          string
	Status           string
	Error            string
	Requests         int64
	Successes        int64
	Failures         int64
	ErrorRate        string
	RPS              string
	DataBytes        string
	P50              string
	P90              string
	P95              string
	P99              string
	StatusCodes      []statusCodeRow
	Errors           []errorRow
	AssertionResults []assertionDetailRow
	AssertionsPassed bool
}

type statusCodeRow struct {
	Code  int
	Count int64
}

type errorRow struct {
	Message string
	Count   int64
}

type assertionDetailRow struct {
	Metric    string
	Expected  string
	Actual    string
	Passed    bool
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	records := s.runner.Results().All()

	rows := make([]historyRow, 0, len(records))
	for _, rec := range records {
		rows = append(rows, toHistoryRow(rec))
	}

	s.templates.renderPage(w, "history", PageData{
		Title:     "History",
		ActiveNav: "history",
		Username:  usernameFromContext(r.Context()),
		Data:      rows,
	})
}

func (s *Server) handleHistoryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, apiError("id is required"))
		return
	}

	rec := s.runner.Results().ByID(id)
	if rec == nil {
		writeJSON(w, http.StatusNotFound, apiError("run not found"))
		return
	}

	writeJSON(w, http.StatusOK, toRunDetail(rec))
}

func toHistoryRow(rec *runner.RunRecord) historyRow {
	row := historyRow{
		ID:       rec.ID,
		SpecURL:  rec.SpecURL,
		Profile:  rec.Profile,
		Workers:  rec.Workers,
		Duration: rec.Duration,
		Status:   string(rec.Status),
	}

	if !rec.StartedAt.IsZero() {
		row.StartedAt = rec.StartedAt.Format("Jan 2, 15:04:05")
	}

	end := rec.EndedAt
	if end.IsZero() {
		end = time.Now()
	}
	row.Elapsed = end.Sub(rec.StartedAt).Round(time.Second).String()

	if rec.Snapshot != nil {
		snap := rec.Snapshot
		row.Requests = snap.TotalRequests
		row.Successes = snap.TotalSuccesses
		row.Failures = snap.TotalFailures
		row.ErrorRate = formatRate(snap)
		row.RPS = formatRPS(snap.RPS)
	} else {
		row.ErrorRate = "—"
		row.RPS = "—"
	}

	return row
}

func toRunDetail(rec *runner.RunRecord) runDetail {
	d := runDetail{
		ID:       rec.ID,
		SpecURL:  rec.SpecURL,
		Profile:  rec.Profile,
		Workers:  rec.Workers,
		Duration: rec.Duration,
		Status:   string(rec.Status),
		Error:    rec.Error,
	}

	if !rec.StartedAt.IsZero() {
		d.StartedAt = rec.StartedAt.Format("Jan 2, 2006 15:04:05")
	}
	if !rec.EndedAt.IsZero() {
		d.EndedAt = rec.EndedAt.Format("Jan 2, 2006 15:04:05")
		d.Elapsed = rec.EndedAt.Sub(rec.StartedAt).Round(time.Millisecond).String()
	} else {
		d.Elapsed = time.Since(rec.StartedAt).Round(time.Second).String()
	}

	if rec.Snapshot != nil {
		snap := rec.Snapshot
		d.Requests = snap.TotalRequests
		d.Successes = snap.TotalSuccesses
		d.Failures = snap.TotalFailures
		d.ErrorRate = formatRate(snap)
		d.RPS = formatRPS(snap.RPS)
		d.DataBytes = formatBytes(snap.TotalBytes)

		codes := make([]statusCodeRow, 0, len(snap.StatusCodes))
		for code, count := range snap.StatusCodes {
			codes = append(codes, statusCodeRow{Code: code, Count: count})
		}
		sort.Slice(codes, func(i, j int) bool { return codes[i].Code < codes[j].Code })
		d.StatusCodes = codes

		errs := make([]errorRow, 0, len(snap.Errors))
		for msg, count := range snap.Errors {
			errs = append(errs, errorRow{Message: msg, Count: count})
		}
		sort.Slice(errs, func(i, j int) bool { return errs[i].Count > errs[j].Count })
		d.Errors = errs
	}

	if rec.Percentiles != nil {
		p := rec.Percentiles
		d.P50 = p.P50.Round(time.Millisecond).String()
		d.P90 = p.P90.Round(time.Millisecond).String()
		d.P95 = p.P95.Round(time.Millisecond).String()
		d.P99 = p.P99.Round(time.Millisecond).String()
	}

	d.AssertionsPassed = rec.AssertionsPassed
	for _, ar := range rec.AssertionResults {
		d.AssertionResults = append(d.AssertionResults, assertionDetailRow{
			Metric:   formatMetricLabel(ar.Metric),
			Expected: formatExpected(ar.Operator, ar.Threshold, ar.Metric),
			Actual:   formatActualValue(ar.Actual, ar.Metric),
			Passed:   ar.Passed,
		})
	}

	return d
}

func formatMetricLabel(metric string) string {
	labels := map[string]string{
		"p50_latency":    "P50 Latency",
		"p90_latency":    "P90 Latency",
		"p95_latency":    "P95 Latency",
		"p99_latency":    "P99 Latency",
		"avg_latency":    "Avg Latency",
		"max_latency":    "Max Latency",
		"error_rate":     "Error Rate",
		"success_rate":   "Success Rate",
		"rps":            "RPS",
		"total_requests": "Total Requests",
		"total_errors":   "Total Errors",
	}
	if label, ok := labels[metric]; ok {
		return label
	}
	return strings.ReplaceAll(metric, "_", " ")
}

func formatExpected(operator string, threshold float64, metric string) string {
	ops := map[string]string{
		"less_than":            "<",
		"less_than_or_equal":   "≤",
		"greater_than":         ">",
		"greater_than_or_equal":"≥",
		"equal":                "=",
	}
	op := operator
	if sym, ok := ops[operator]; ok {
		op = sym
	}
	return fmt.Sprintf("%s %s", op, formatActualValue(threshold, metric))
}

func formatActualValue(value float64, metric string) string {
	switch {
	case strings.HasSuffix(metric, "_latency"):
		return fmt.Sprintf("%.0fms", value)
	case strings.HasSuffix(metric, "_rate"):
		return fmt.Sprintf("%.2f%%", value)
	case metric == "rps":
		return fmt.Sprintf("%.2f req/s", value)
	default:
		return fmt.Sprintf("%.0f", value)
	}
}

func formatRate(snap *loader.MetricsSnapshot) string {
	if snap.TotalRequests == 0 {
		return "0.00%"
	}
	rate := float64(snap.TotalFailures) / float64(snap.TotalRequests) * 100
	return fmt.Sprintf("%.2f%%", rate)
}

func formatRPS(rps float64) string {
	return fmt.Sprintf("%.2f", rps)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
