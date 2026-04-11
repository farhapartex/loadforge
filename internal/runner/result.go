package runner

import (
	"encoding/json"
	"log"
	"os"
	"sort"
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

type LatencyPercentiles struct {
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
}

type RunRecord struct {
	ID          string
	SpecURL     string
	Profile     string
	Workers     int
	Duration    string
	StartedAt   time.Time
	EndedAt     time.Time
	Status      RunStatus
	Error       string
	Snapshot    *loader.MetricsSnapshot
	Percentiles *LatencyPercentiles
}

type ResultStore struct {
	mu          sync.RWMutex
	records     []*RunRecord
	historyFile string
}

func newResultStore(historyFile string) *ResultStore {
	rs := &ResultStore{historyFile: historyFile}
	if historyFile != "" {
		rs.load()
	}
	return rs
}

func (rs *ResultStore) add(r *RunRecord) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.records = append([]*RunRecord{r}, rs.records...)
	if len(rs.records) > maxHistory {
		rs.records = rs.records[:maxHistory]
	}

	if rs.historyFile != "" {
		rs.save()
	}
}

func (rs *ResultStore) All() []*RunRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	out := make([]*RunRecord, len(rs.records))
	copy(out, rs.records)
	return out
}

func (rs *ResultStore) ByID(id string) *RunRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for _, r := range rs.records {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (rs *ResultStore) Latest() *RunRecord {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if len(rs.records) == 0 {
		return nil
	}
	return rs.records[0]
}

func (rs *ResultStore) save() {
	data, err := json.MarshalIndent(rs.records, "", "  ")
	if err != nil {
		log.Printf("WARN history: marshal failed: %v", err)
		return
	}
	if err := os.WriteFile(rs.historyFile, data, 0644); err != nil {
		log.Printf("WARN history: write failed: %v", err)
	}
}

func (rs *ResultStore) load() {
	data, err := os.ReadFile(rs.historyFile)
	if err != nil {
		return
	}
	var records []*RunRecord
	if err := json.Unmarshal(data, &records); err != nil {
		log.Printf("WARN history: parse failed: %v", err)
		return
	}
	rs.records = records
}

func ComputePercentiles(snap *loader.MetricsSnapshot) *LatencyPercentiles {
	if len(snap.Latencies) == 0 {
		return &LatencyPercentiles{}
	}

	sorted := make([]time.Duration, len(snap.Latencies))
	copy(sorted, snap.Latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p := &LatencyPercentiles{
		P50: percentileAt(sorted, 50),
		P90: percentileAt(sorted, 90),
		P95: percentileAt(sorted, 95),
		P99: percentileAt(sorted, 99),
	}

	snap.Latencies = nil
	return p
}

func percentileAt(sorted []time.Duration, n int) time.Duration {
	idx := int(float64(len(sorted)-1) * float64(n) / 100)
	return sorted[idx]
}
