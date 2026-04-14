package web

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/farhapartex/loadforge/internal/config"
)

type WebConfig struct {
	Addr              string             `yaml:"addr"`
	Username          string             `yaml:"username"`
	Password          string             `yaml:"password"`
	PasswordChanged   bool               `yaml:"password_changed"`
	SessionTTL        string             `yaml:"session_ttl"`
	LogFile           string             `yaml:"log_file"`
	HistoryFile       string             `yaml:"history_file"`
	DefaultAssertions []config.Assertion `yaml:"assertions,omitempty"`
}

func (c *WebConfig) parsedSessionTTL() time.Duration {
	if d, err := time.ParseDuration(c.SessionTTL); err == nil {
		return d
	}
	return 24 * time.Hour
}

func (c *WebConfig) save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("commit config: %w", err)
	}
	return nil
}

func loadWebConfig(path string) (*WebConfig, error) {
	cfg := defaultWebConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if len(cfg.DefaultAssertions) == 0 {
		cfg.DefaultAssertions = defaultAssertions()
		_ = cfg.save(path)
	}

	return cfg, nil
}

func defaultWebConfig() *WebConfig {
	logFile := "load_forge.logs"
	historyFile := "load_forge_history.json"

	if dir, err := loadForgeDir(); err == nil {
		logFile = dir + "/load_forge.logs"
		historyFile = dir + "/load_forge_history.json"
	}

	return &WebConfig{
		Addr:              ":8080",
		Username:          "admin",
		Password:          "admin",
		SessionTTL:        "24h",
		LogFile:           logFile,
		HistoryFile:       historyFile,
		DefaultAssertions: defaultAssertions(),
	}
}

func defaultAssertions() []config.Assertion {
	return []config.Assertion{
		{Metric: "p95_latency",  Operator: "less_than",            Value: 500,  Enabled: true},
		{Metric: "p99_latency",  Operator: "less_than",            Value: 1000, Enabled: true},
		{Metric: "error_rate",   Operator: "less_than",            Value: 1.0,  Enabled: true},
		{Metric: "success_rate", Operator: "greater_than_or_equal", Value: 99.0, Enabled: false},
		{Metric: "rps",          Operator: "greater_than",         Value: 10,   Enabled: false},
	}
}
