package web

import (
	"fmt"

	"github.com/farhapartex/loadforge/internal/openapi"
	"github.com/farhapartex/loadforge/internal/postman"
	"github.com/farhapartex/loadforge/internal/specloader"
)

type loaderRegistry struct {
	loaders map[string]specloader.Loader
}

func newLoaderRegistry() *loaderRegistry {
	r := &loaderRegistry{loaders: make(map[string]specloader.Loader)}
	r.register(openapi.NewLoader())
	r.register(postman.NewLoader())
	return r
}

func (r *loaderRegistry) register(l specloader.Loader) {
	r.loaders[l.Name()] = l
}

func (r *loaderRegistry) get(sourceType string) (specloader.Loader, error) {
	l, ok := r.loaders[sourceType]
	if !ok {
		return nil, fmt.Errorf("unknown source type %q", sourceType)
	}
	return l, nil
}
