package web

import (
	"sync"
	"time"
)

type RunStats struct {
	mu          sync.RWMutex
	totalRuns   int
	lastStatus  string
	lastRunAt   time.Time
	lastConfig  string
	activeTests int
}

type RunStatsSnapshot struct {
	TotalRuns   int
	LastStatus  string
	LastRunAt   string
	LastConfig  string
	ActiveTests int
}

func newRunStats() *RunStats {
	return &RunStats{}
}

func (s *RunStats) snapshot() RunStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastRunAt := "Never"
	if !s.lastRunAt.IsZero() {
		lastRunAt = s.lastRunAt.Format("Jan 2, 15:04")
	}
	lastStatus := s.lastStatus
	if lastStatus == "" {
		lastStatus = "N/A"
	}
	lastConfig := s.lastConfig
	if lastConfig == "" {
		lastConfig = "N/A"
	}
	return RunStatsSnapshot{
		TotalRuns:   s.totalRuns,
		LastStatus:  lastStatus,
		LastRunAt:   lastRunAt,
		LastConfig:  lastConfig,
		ActiveTests: s.activeTests,
	}
}
