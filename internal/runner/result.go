package runner

import (
	"sync"
	"time"

	"github.com/farhapartex/loadforge/internal/loader"
)

const maxHistory = 100

type RunStatus string

const (
	StatusRunning   RunStatus = "running"
	StatusCompleted RunStatus = "completed"
	StatusFailed    RunStatus = "failed"
	StatusStopped   RunStatus = "stopped"
)

type RunRecord struct {
	ID        string
	SpecURL   string
	Profile   string
	Workers   int
	Duration  string
	StartedAt time.Time
	EndedAt   time.Time
	Status    RunStatus
	Error     string
	Snapshot  *loader.MetricsSnapshot
}

type ResultStore struct {
	mu      sync.RWMutex
	records []*RunRecord
}

func newResultStore() *ResultStore {
	return &ResultStore{}
}

func (rs *ResultStore) add(r *RunRecord) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.records = append([]*RunRecord{r}, rs.records...)
	if len(rs.records) > maxHistory {
		rs.records = rs.records[:maxHistory]
	}
}

func (rs *ResultStore) All() []*RunRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	out := make([]*RunRecord, len(rs.records))
	copy(out, rs.records)
	return out
}

func (rs *ResultStore) Latest() *RunRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if len(rs.records) == 0 {
		return nil
	}
	return rs.records[0]
}
