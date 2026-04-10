package web

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/farhapartex/loadforge/internal/loader"
	"github.com/farhapartex/loadforge/internal/runner"
)

type historyRow struct {
	ID         string
	SpecURL    string
	StartedAt  string
	Duration   string
	Status     string
	Requests   int64
	Successes  int64
	Failures   int64
	ErrorRate  string
	RPS        string
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	records := s.runner.Results().All()

	rows := make([]historyRow, 0, len(records))
	for _, rec := range records {
		rows = append(rows, toHistoryRow(rec))
	}

	sort.Slice(rows, func(i, j int) bool {
		return records[i].StartedAt.After(records[j].StartedAt)
	})

	s.templates.renderPage(w, "history", PageData{
		Title:     "History",
		ActiveNav: "history",
		Username:  usernameFromContext(r.Context()),
		Data:      rows,
	})
}

func toHistoryRow(rec *runner.RunRecord) historyRow {
	row := historyRow{
		ID:      rec.ID,
		SpecURL: rec.SpecURL,
		Status:  string(rec.Status),
	}

	if !rec.StartedAt.IsZero() {
		row.StartedAt = rec.StartedAt.Format("Jan 2, 15:04:05")
	}

	end := rec.EndedAt
	if end.IsZero() {
		end = time.Now()
	}
	row.Duration = end.Sub(rec.StartedAt).Round(time.Second).String()

	if rec.Snapshot != nil {
		snap := rec.Snapshot
		row.Requests = snap.TotalRequests
		row.Successes = snap.TotalSuccesses
		row.Failures = snap.TotalFailures
		row.ErrorRate = formatRate(snap)
		row.RPS = formatRPS(snap.RPS)
	}

	return row
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
