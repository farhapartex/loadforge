package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Name      string     `yaml:"name"`
	BaseURL   string     `yaml:"base_url"`
	Scenarios []Scenario `yaml:"scenarios"`
	Load      LoadConfig `yaml:"load"`
}

type Scenario struct {
	Name   string `yaml:"name"`
	Weight int    `yaml:"weight"`
	Steps  []Step `yaml:"steps"`
}

type Step struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    *Body             `yaml:"body"`
	Auth    *Auth             `yaml:"auth"`
	Options *RequestOptions   `yaml:"options"`
	Think   string            `yaml:"think"`
}

type Body struct {
	Raw  string            `yaml:"raw"`
	JSON any              `yaml:"json"`
	Form map[string]string `yaml:"form"`
}

type Auth struct {
	Basic  *BasicAuth  `yaml:"basic"`
	Bearer string      `yaml:"bearer"`
	Header *HeaderAuth `yaml:"header"`
}

type RequestOptions struct {
	Timeout         string `yaml:"timeout"`
	FollowRedirects *bool  `yaml:"follow_redirects"`
	TLSSkipVerify   bool   `yaml:"tls_skip_verify"`
	HTTP2           *bool  `yaml:"http2"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type HeaderAuth struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// LoadConfig defines how load is applied over time.
type LoadConfig struct {
	Profile     string        `yaml:"profile"`
	Duration    string        `yaml:"duration"`
	Workers     int           `yaml:"workers"`
	RampUp      *RampUpConfig `yaml:"ramp_up"`
	Step        *StepConfig   `yaml:"step"`
	Spike       *SpikeConfig  `yaml:"spike"`
	MaxRequests int           `yaml:"max_requests"`
}

type Constant struct {
	Workers  int    `yaml:"workers"`
	Duration string `yaml:"duration"`
	Requests int    `yaml:"requests"`
}

type RampUpConfig struct {
	StartWorkers int    `yaml:"start_workers"`
	EndWorkers   int    `yaml:"end_workers"`
	Duration     string `yaml:"duration"`
}

type StepConfig struct {
	StartWorkers int    `yaml:"start_workers"`
	StepSize     int    `yaml:"step_size"`
	StepDuration string `yaml:"step_duration"`
	MaxWorkers   int    `yaml:"max_workers"`
}

type SpikeConfig struct {
	BaseWorkers   int    `yaml:"base_workers"`
	SpikeWorkers  int    `yaml:"spike_workers"`
	SpikeDuration string `yaml:"spike_duration"`
	SpikeEvery    string `yaml:"spike_every"`
}

// LoadFromFile reads a YAML config file from disk, parses and validates it.
// Used by the CLI.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Default returns a base Config with sensible load defaults and no scenarios.
// The web layer and OpenAPI generator use this as a starting point, populate
// Scenarios and BaseURL, then call Validate before running.
func Default() *Config {
	return &Config{
		Name:      "load-test",
		BaseURL:   "",
		Scenarios: []Scenario{},
		Load: LoadConfig{
			Profile:  "constant",
			Duration: "30s",
			Workers:  10,
		},
	}
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("config must have a name")
	}

	if len(c.Scenarios) == 0 {
		return fmt.Errorf("config must have at least one scenario")
	}

	for idx, scenario := range c.Scenarios {
		if len(scenario.Steps) == 0 {
			return fmt.Errorf("scenario[%d] %q must have at least one step", idx, scenario.Name)
		}

		for jdx, step := range scenario.Steps {
			if err := validateStep(idx, jdx, step); err != nil {
				return err
			}
		}
	}

	if err := c.Load.Validate(); err != nil {
		return fmt.Errorf("invalid load profile: %w", err)
	}

	return nil
}

func validateStep(scenarioIdx, stepIdx int, step Step) error {
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true,
		"PATCH": true, "DELETE": true, "HEAD": true,
	}

	if step.URL == "" {
		return fmt.Errorf("scenario[%d].step[%d] must have a url", scenarioIdx, stepIdx)
	}

	if step.Method == "" {
		return fmt.Errorf("scenario[%d].step[%d] must have a method", scenarioIdx, stepIdx)
	}

	if !validMethods[step.Method] {
		return fmt.Errorf("scenario[%d].step[%d] has invalid method %q", scenarioIdx, stepIdx, step.Method)
	}

	if step.Options != nil && step.Options.Timeout != "" {
		if _, err := time.ParseDuration(step.Options.Timeout); err != nil {
			return fmt.Errorf("scenario[%d].step[%d] has invalid timeout %q: %w",
				scenarioIdx, stepIdx, step.Options.Timeout, err)
		}
	}

	if step.Think != "" {
		if _, err := time.ParseDuration(step.Think); err != nil {
			return fmt.Errorf("scenario[%d].step[%d] has invalid think time %q: %w",
				scenarioIdx, stepIdx, step.Think, err)
		}
	}

	return nil
}

func (lc *LoadConfig) Validate() error {
	validProfiles := map[string]bool{
		"constant": true, "ramp": true, "step": true, "spike": true,
	}

	if lc.Profile == "" {
		lc.Profile = "constant"
	}

	if !validProfiles[lc.Profile] {
		return fmt.Errorf("unknown profile %q, must be one of: constant, ramp, step, spike", lc.Profile)
	}

	if lc.Duration == "" && lc.MaxRequests == 0 {
		return fmt.Errorf("load must have either duration or max_requests set")
	}

	if lc.Duration != "" {
		if _, err := time.ParseDuration(lc.Duration); err != nil {
			return fmt.Errorf("invalid duration %q: %w", lc.Duration, err)
		}
	}

	switch lc.Profile {
	case "constant":
		if lc.Workers <= 0 {
			return fmt.Errorf("constant profile requires workers > 0")
		}
	case "ramp":
		if lc.RampUp == nil {
			return fmt.Errorf("ramp profile requires ramp_up config")
		}
		if lc.RampUp.EndWorkers <= 0 {
			return fmt.Errorf("ramp_up.end_workers must be > 0")
		}
		if lc.RampUp.Duration == "" {
			return fmt.Errorf("ramp_up.duration is required")
		}
		if _, err := time.ParseDuration(lc.RampUp.Duration); err != nil {
			return fmt.Errorf("ramp_up.duration invalid: %w", err)
		}
	case "step":
		if lc.Step == nil {
			return fmt.Errorf("step profile requires step config")
		}
		if lc.Step.StepSize <= 0 {
			return fmt.Errorf("step.step_size must be > 0")
		}
		if lc.Step.MaxWorkers <= 0 {
			return fmt.Errorf("step.max_workers must be > 0")
		}
		if lc.Step.StepDuration == "" {
			return fmt.Errorf("step.step_duration is required")
		}
		if _, err := time.ParseDuration(lc.Step.StepDuration); err != nil {
			return fmt.Errorf("step.step_duration invalid: %w", err)
		}
	case "spike":
		if lc.Spike == nil {
			return fmt.Errorf("spike profile requires spike config")
		}
		if lc.Spike.BaseWorkers <= 0 {
			return fmt.Errorf("spike.base_workers must be > 0")
		}
		if lc.Spike.SpikeWorkers <= 0 {
			return fmt.Errorf("spike.spike_workers must be > 0")
		}
		if lc.Spike.SpikeDuration == "" {
			return fmt.Errorf("spike.spike_duration is required")
		}
		if lc.Spike.SpikeEvery == "" {
			return fmt.Errorf("spike.spike_every is required")
		}
	}

	return nil
}
