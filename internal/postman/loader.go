package postman

import (
	"fmt"

	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/specloader"
)

// PostmanLoader implements specloader.Loader for Postman collection v2.1 files.
// This is a reserved stub — full implementation is planned for a future release.
type PostmanLoader struct{}

func NewLoader() *PostmanLoader { return &PostmanLoader{} }

func (l *PostmanLoader) Name() string { return "postman" }

func (l *PostmanLoader) Load(_ specloader.Input) (*config.Config, error) {
	return nil, fmt.Errorf("postman collection support is not yet implemented")
}
