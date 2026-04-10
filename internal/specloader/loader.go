package specloader

import "github.com/farhapartex/loadforge/internal/config"

// Loader converts an external spec source into a load-test config.
// Both OpenAPI and Postman implement this interface — the web layer
// picks the right implementation based on the user's source selection.
type Loader interface {
	// Name returns a human-readable identifier used in logs and history.
	Name() string
	// Load parses the input and returns a ready-to-run config.
	Load(input Input) (*config.Config, error)
}

// Input carries everything a Loader needs. URL-based loaders use URL+Token;
// file-based loaders use Data+Filename. Load profile overrides are applied
// after the loader builds its base config.
type Input struct {
	// URL-based source (OpenAPI spec URL)
	URL   string
	Token string

	// File-based source (Postman collection upload)
	Data     []byte
	Filename string

	// Load profile overrides — zero values mean "use loader defaults"
	Workers  int
	Duration string
	Profile  string
}
