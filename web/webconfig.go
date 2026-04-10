package web

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type WebConfig struct {
	Addr       string `yaml:"addr"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	SessionTTL string `yaml:"session_ttl"`
	LogFile    string `yaml:"log_file"`
}

func (c *WebConfig) parsedSessionTTL() time.Duration {
	if d, err := time.ParseDuration(c.SessionTTL); err == nil {
		return d
	}
	return 24 * time.Hour
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

	return cfg, nil
}

func defaultWebConfig() *WebConfig {
	return &WebConfig{
		Addr:       ":8080",
		Username:   "admin",
		Password:   "admin",
		SessionTTL: "24h",
		LogFile:    "load_forge.logs",
	}
}
