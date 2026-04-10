package openapi

import (
	"fmt"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/specloader"
)

// OpenAPILoader implements specloader.Loader for OpenAPI 3.x / Swagger 2.x specs.
type OpenAPILoader struct{}

func NewLoader() *OpenAPILoader { return &OpenAPILoader{} }

func (l *OpenAPILoader) Name() string { return "openapi" }

func (l *OpenAPILoader) Load(input specloader.Input) (*config.Config, error) {
	if input.URL == "" && len(input.Data) == 0 {
		return nil, fmt.Errorf("openapi loader requires either a URL or raw spec data")
	}

	var data []byte

	if len(input.Data) > 0 {
		data = input.Data
	} else {
		var err error
		data, _, err = Fetch(input.URL, input.Token)
		if err != nil {
			return nil, fmt.Errorf("fetch spec: %w", err)
		}
	}

	spec, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}

	ops := Extract(spec)
	if len(ops) == 0 {
		return nil, fmt.Errorf("no operations found in spec")
	}

	opts := GenerateOptions{
		Token:   input.Token,
		Profile: "constant",
	}
	if input.Profile != "" {
		opts.Profile = input.Profile
	}
	if input.Workers > 0 {
		opts.Workers = input.Workers
	}
	if input.Duration != "" {
		opts.Duration = input.Duration
	}

	cfg, err := Generate(ops, spec.BaseURL, opts)
	if err != nil {
		return nil, fmt.Errorf("generate config: %w", err)
	}

	return cfg, nil
}
