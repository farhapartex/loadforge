package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Name        string      `yaml:"name"`
	BaseURL     string      `yaml:"base_url"`
	Scenarios   []Scenario  `yaml:"scenarios"`
	LoadProfile LoadProfile `yaml:"load_profile"`
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
	JSON interface{}       `yaml:"json"`
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
	Pasword  string `yaml:"password"`
}

type HeaderAuth struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// LoadProfile define how load is applied over time
type LoadProfile struct {
	Type     string       `yaml:"type"`
	Constant *Constant    `yaml:"constant"`
	Ramp     *Ramp        `yaml:"ramp"`
	Step     *StepProfile `yaml:"step_profile"`
	Spike    *Spike       `yaml:"spike"`
}

// Constant keeps a fixed number of workers for a the entire duration
type Constant struct {
	Workers  int    `yaml:"workers"`
	Duration string `yaml:"duration"`
	Requests int    `yaml:"requests"`
}

type Ramp struct {
	Target       int    `yaml:"target"`
	RampDuration string `yaml:"ramp_duration"`
	HoldDuration string `yaml:"hold_duration"`
}

// Spike runs at baseline, suddenly spikes then returns to baseline
type Spike struct {
	Baseline      int    `yaml:"baseline"`
	Peak          int    `yaml:"peak"`
	BaselinePre   string `yaml:"baseline_pre"`
	SpikeDuration string `yaml:"spike_duration"`
	BaselinePost  string `yaml:"baseline_post"`
}

// StepProfile increases workers in stages
type StepProfile struct {
	Start      int    `yaml:"start"`
	StepSize   int    `yaml:"step_size"`
	StepEvery  string `yaml:"step_every"`
	MaxWorkers int    `yaml:"max_workers"`
	Duration   string `yaml:"duration"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("Invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("Config must have a name")
	}

	if len(c.Scenarios) == 0 {
		return fmt.Errorf("Config must have at least one scenario")
	}

	for idx, scenario := range c.Scenarios {
		if len(scenario.Steps) == 0 {
			return fmt.Errorf("Scenario[%d] %q must have at least one step", idx, scenario.Name)
		}

		for jdx, step := range scenario.Steps {
			if err := validateStep(idx, jdx, step); err != nil {
				return err
			}
		}
	}

	if err := c.LoadProfile.Validate(); err != nil {
		return fmt.Errorf("invalid load_profile: %w", err)
	}

	return nil
}

func validateStep(scenarioIdx int, stepIdx int, step Step) error {
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
		"HEAD":   true,
	}

	if step.URL == "" {
		return fmt.Errorf("scenario[%d].step[%d] must have a url", scenarioIdx, stepIdx)
	}

	if step.Method == "" {
		return fmt.Errorf("scenario[%d].step[%d] must have a url", scenarioIdx, stepIdx)
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

func (lp *LoadProfile) Validate() error {
	validTypes := map[string]bool{
		"constant": true,
		"ramp":     true,
		"step":     true,
		"spike":    true,
	}

	if lp.Type == "" {
		return fmt.Errorf(("load_profile.type is required"))
	}

	if !validTypes[lp.Type] {
		return fmt.Errorf("unknown load_profile type %q, must be one: constant, ramp, step and spikle", lp.Type)
	}

	switch lp.Type {
	case "constant":
		if lp.Constant == nil {
			return fmt.Errorf("load_profile.constant config is required when type is constant")
		}

		if lp.Constant.Workers <= 0 {
			return fmt.Errorf("load_profile.constant.workers must be > 0")
		}

		if lp.Constant.Duration == "" && lp.Constant.Requests <= 0 {
			return fmt.Errorf("load_profile.constant musthave either duration or requests set")
		}

		if lp.Constant.Duration != "" {
			if _, err := time.ParseDuration(lp.Constant.Duration); err != nil {
				return fmt.Errorf("invalid constant.duration: %w", err)
			}
		}

	case "ramp":
		if lp.Ramp == nil {
			return fmt.Errorf("load_profile.ramp config is required when type is ramp")
		}
		if lp.Ramp.Target <= 0 {
			return fmt.Errorf("load_profile.ramp.target must be > 0")
		}
		if _, err := time.ParseDuration(lp.Ramp.RampDuration); err != nil {
			return fmt.Errorf("invalid ramp.ramp_duration: %w", err)
		}
		if _, err := time.ParseDuration(lp.Ramp.HoldDuration); err != nil {
			return fmt.Errorf("invalid ramp.hold_duration: %w", err)
		}

	case "step":
		if lp.Step == nil {
			return fmt.Errorf("load_profile.step_profile config is required when type is step")
		}
		if lp.Step.Start <= 0 {
			return fmt.Errorf("load_profile.step_profile.start must be > 0")
		}
		if lp.Step.StepSize <= 0 {
			return fmt.Errorf("load_profile.step_profile.step_size must be > 0")
		}
		if lp.Step.MaxWorkers <= 0 {
			return fmt.Errorf("load_profile.step_profile.max_workers must be > 0")
		}
		if _, err := time.ParseDuration(lp.Step.StepEvery); err != nil {
			return fmt.Errorf("invalid step_profile.step_every: %w", err)
		}
		if _, err := time.ParseDuration(lp.Step.Duration); err != nil {
			return fmt.Errorf("invalid step_profile.duration: %w", err)
		}

	case "spike":
		if lp.Spike == nil {
			return fmt.Errorf("load_profile.spike config is required when type is spike")
		}
		if lp.Spike.Baseline <= 0 {
			return fmt.Errorf("load_profile.spike.baseline must be > 0")
		}
		if lp.Spike.Peak <= lp.Spike.Baseline {
			return fmt.Errorf("load_profile.spike.peak must be greater than baseline")
		}
		for _, d := range []string{lp.Spike.BaselinePre, lp.Spike.SpikeDuration, lp.Spike.BaselinePost} {
			if _, err := time.ParseDuration(d); err != nil {
				return fmt.Errorf("invalid spike duration %q: %w", d, err)
			}
		}
	}

	return nil
}
