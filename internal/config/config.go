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
