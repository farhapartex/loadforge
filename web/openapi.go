package web

import (
	"github.com/farhapartex/loadforge/internal/config"
	"github.com/farhapartex/loadforge/internal/openapi"
)

type openapiClient struct{}

func newOpenapiClient() *openapiClient { return &openapiClient{} }

func (c *openapiClient) Fetch(specURL, token string) ([]byte, string, error) {
	return openapi.Fetch(specURL, token)
}

func (c *openapiClient) Parse(data []byte) (*openapi.Spec, error) {
	return openapi.Parse(data)
}

func (c *openapiClient) Extract(spec *openapi.Spec) []openapi.Operation {
	return openapi.Extract(spec)
}

func (c *openapiClient) Generate(ops []openapi.Operation, baseURL, token string) (*config.Config, error) {
	return openapi.Generate(ops, baseURL, openapi.GenerateOptions{
		Token:   token,
		Profile: "constant",
	})
}
