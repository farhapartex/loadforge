package openapi

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
)

type GenerateOptions struct {
	Token    string
	Workers  int
	Duration string
	Profile  string
	OneScenario bool
}

func Generate(ops []Operation, baseURL string, opts GenerateOptions) (*config.Config, error) {
	if len(ops) == 0 {
		return nil, fmt.Errorf("no operations extracted from spec")
	}

	cfg := config.Default()
	cfg.BaseURL = baseURL

	if opts.Workers > 0 {
		cfg.Load.Workers = opts.Workers
	}
	if opts.Duration != "" {
		cfg.Load.Duration = opts.Duration
	}
	if opts.Profile != "" {
		cfg.Load.Profile = opts.Profile
	}

	if opts.OneScenario {
		scenario := buildScenario("all-endpoints", ops, opts.Token)
		cfg.Scenarios = []config.Scenario{scenario}
	} else {
		cfg.Scenarios = buildGroupedScenarios(ops, opts.Token)
	}

	if len(cfg.Scenarios) == 0 {
		return nil, fmt.Errorf("could not build any scenarios from extracted operations")
	}

	applyProfileDefaults(&cfg.Load)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("generated config is invalid: %w", err)
	}

	return cfg, nil
}

func buildGroupedScenarios(ops []Operation, token string) []config.Scenario {
	groups := make(map[string][]Operation)
	var order []string

	for _, op := range ops {
		tag := firstTag(op.Tags)
		if _, seen := groups[tag]; !seen {
			order = append(order, tag)
		}
		groups[tag] = append(groups[tag], op)
	}

	sort.Strings(order)

	var scenarios []config.Scenario
	for _, tag := range order {
		groupOps := groups[tag]
		scenario := buildScenario(tag, groupOps, token)
		if len(scenario.Steps) > 0 {
			scenarios = append(scenarios, scenario)
		}
	}

	return scenarios
}

func buildScenario(name string, ops []Operation, token string) config.Scenario {
	scenario := config.Scenario{
		Name:   name,
		Weight: 1,
	}

	for _, op := range ops {
		step := buildStep(op, token)
		scenario.Steps = append(scenario.Steps, step)
	}

	return scenario
}

func buildStep(op Operation, token string) config.Step {
	resolvedPath := ResolvePath(op.Path, op.PathParams)

	step := config.Step{
		Name:   stepName(op),
		Method: op.Method,
		URL:    resolvedPath,
	}

	if len(op.QueryParams) > 0 {
		step.URL = resolvedPath + buildQueryString(op.QueryParams)
	}

	if len(op.HeaderParams) > 0 {
		step.Headers = make(map[string]string)
		for _, p := range op.HeaderParams {
			step.Headers[p.Name] = paramValueString(p)
		}
	}

	if op.Body != nil && len(op.Body) > 0 {
		step.Body = &config.Body{JSON: op.Body}
		if step.Headers == nil {
			step.Headers = make(map[string]string)
		}
		if op.ContentType != "" && op.ContentType != "application/json" {
			step.Headers["Content-Type"] = op.ContentType
		}
	}

	if token != "" {
		step.Auth = &config.Auth{Bearer: token}
	}

	return step
}

func stepName(op Operation) string {
	if op.OperationID != "" {
		return op.OperationID
	}
	if op.Summary != "" {
		return op.Summary
	}
	return fmt.Sprintf("%s %s", op.Method, op.Path)
}

func firstTag(tags []string) string {
	if len(tags) > 0 && tags[0] != "" {
		return tags[0]
	}
	return "default"
}

func applyProfileDefaults(lc *config.LoadConfig) {
	totalDur, err := time.ParseDuration(lc.Duration)
	if err != nil {
		totalDur = 30 * time.Second
	}
	workers := lc.Workers
	if workers <= 0 {
		workers = 10
	}

	switch lc.Profile {
	case "ramp":
		if lc.RampUp == nil {
			rampDur := totalDur / 2
			if rampDur < time.Second {
				rampDur = time.Second
			}
			lc.RampUp = &config.RampUpConfig{
				StartWorkers: 1,
				EndWorkers:   workers,
				Duration:     rampDur.String(),
			}
		}

	case "step":
		if lc.Step == nil {
			stepSize := workers / 5
			if stepSize < 1 {
				stepSize = 1
			}
			steps := workers / stepSize
			if steps < 1 {
				steps = 1
			}
			stepDur := totalDur / time.Duration(steps)
			if stepDur < time.Second {
				stepDur = time.Second
			}
			lc.Step = &config.StepConfig{
				StartWorkers: 1,
				StepSize:     stepSize,
				StepDuration: stepDur.String(),
				MaxWorkers:   workers,
			}
		}

	case "spike":
		if lc.Spike == nil {
			base := workers / 4
			if base < 1 {
				base = 1
			}
			spikeDur := 10 * time.Second
			spikeEvery := 30 * time.Second
			if totalDur < spikeEvery*2 {
				spikeEvery = totalDur / 3
				spikeDur = spikeEvery / 3
			}
			if spikeDur < time.Second {
				spikeDur = time.Second
			}
			if spikeEvery < time.Second {
				spikeEvery = time.Second
			}
			lc.Spike = &config.SpikeConfig{
				BaseWorkers:   base,
				SpikeWorkers:  workers,
				SpikeDuration: spikeDur.String(),
				SpikeEvery:    spikeEvery.String(),
			}
		}
	}
}

func buildQueryString(params []Param) string {
	var parts []string
	for _, p := range params {
		if p.Required {
			parts = append(parts, p.Name+"="+paramValueString(p))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "?" + strings.Join(parts, "&")
}
