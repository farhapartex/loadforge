package openapi

import (
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

type specVersion int

const (
	versionUnknown specVersion = iota
	versionSwagger2
	versionOpenAPI3
)

type Spec struct {
	Version    specVersion
	Title      string
	BaseURL    string
	Paths      map[string]PathItem
	Components map[string]any
	Definitions map[string]any
}

type PathItem struct {
	Operations map[string]RawOp
}

type RawOp struct {
	OperationID string
	Summary     string
	Tags        []string
	Parameters  []RawParam
	RequestBody *RawBody
	Security    []map[string][]string
	Consumes    []string
}

type RawParam struct {
	Name     string
	In       string
	Required bool
	Schema   map[string]any
	Type     string
	Example  any
}

type RawBody struct {
	Required bool
	Content  map[string]RawMediaType
}

type RawMediaType struct {
	Schema map[string]any
}

type rawSpec struct {
	Swagger    string                     `yaml:"swagger"`
	OpenAPI    string                     `yaml:"openapi"`
	Info       map[string]any             `yaml:"info"`
	Host       string                     `yaml:"host"`
	BasePath   string                     `yaml:"basePath"`
	Schemes    []string                   `yaml:"schemes"`
	Servers    []map[string]any           `yaml:"servers"`
	Paths      map[string]map[string]any  `yaml:"paths"`
	Components map[string]any             `yaml:"components"`
	Definitions map[string]any            `yaml:"definitions"`
	Consumes   []string                   `yaml:"consumes"`
}

func Parse(data []byte) (*Spec, error) {
	var raw rawSpec
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}

	version := detectVersion(raw)
	if version == versionUnknown {
		return nil, fmt.Errorf("unrecognized spec format: missing 'openapi' or 'swagger' field")
	}

	spec := &Spec{
		Version:     version,
		Paths:       make(map[string]PathItem),
		Components:  raw.Components,
		Definitions: raw.Definitions,
	}

	if raw.Info != nil {
		if title, ok := raw.Info["title"].(string); ok {
			spec.Title = title
		}
	}

	spec.BaseURL = extractBaseURL(raw, version)

	for path, methods := range raw.Paths {
		item := PathItem{Operations: make(map[string]RawOp)}
		for method, opData := range methods {
			method = strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}
			op, err := parseOperation(opData, raw.Consumes, spec)
			if err != nil {
				continue
			}
			item.Operations[method] = op
		}
		spec.Paths[path] = item
	}

	return spec, nil
}

func detectVersion(raw rawSpec) specVersion {
	if strings.HasPrefix(raw.OpenAPI, "3.") {
		return versionOpenAPI3
	}
	if raw.Swagger == "2.0" || strings.HasPrefix(raw.Swagger, "2.") {
		return versionSwagger2
	}
	return versionUnknown
}

func extractBaseURL(raw rawSpec, version specVersion) string {
	if version == versionOpenAPI3 && len(raw.Servers) > 0 {
		if u, ok := raw.Servers[0]["url"].(string); ok && u != "" {
			return strings.TrimRight(u, "/")
		}
	}

	if version == versionSwagger2 {
		host := raw.Host
		if host == "" {
			return ""
		}
		scheme := "https"
		if len(raw.Schemes) > 0 {
			scheme = raw.Schemes[0]
		}
		base := raw.BasePath
		if base == "/" {
			base = ""
		}
		return scheme + "://" + strings.TrimRight(host+base, "/")
	}

	return ""
}

func parseOperation(opData any, globalConsumes []string, spec *Spec) (RawOp, error) {
	m, ok := toStringMap(opData)
	if !ok {
		return RawOp{}, fmt.Errorf("operation is not a map")
	}

	op := RawOp{}
	op.OperationID, _ = m["operationId"].(string)
	op.Summary, _ = m["summary"].(string)

	if tags, ok := m["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				op.Tags = append(op.Tags, s)
			}
		}
	}

	if params, ok := m["parameters"].([]any); ok {
		for _, p := range params {
			param := resolveParam(p, spec)
			if param != nil {
				op.Parameters = append(op.Parameters, *param)
			}
		}
	}

	if rb, ok := m["requestBody"]; ok {
		op.RequestBody = parseRequestBody(rb, spec)
	}

	if consumes, ok := m["consumes"].([]any); ok {
		for _, c := range consumes {
			if s, ok := c.(string); ok {
				op.Consumes = append(op.Consumes, s)
			}
		}
	}
	if len(op.Consumes) == 0 {
		op.Consumes = globalConsumes
	}

	if sec, ok := m["security"].([]any); ok {
		for _, s := range sec {
			if sm, ok := toStringMap(s); ok {
				entry := make(map[string][]string)
				for k, v := range sm {
					if scopes, ok := v.([]any); ok {
						var ss []string
						for _, sc := range scopes {
							if str, ok := sc.(string); ok {
								ss = append(ss, str)
							}
						}
						entry[k] = ss
					}
				}
				op.Security = append(op.Security, entry)
			}
		}
	}

	return op, nil
}

func resolveParam(p any, spec *Spec) *RawParam {
	m, ok := toStringMap(p)
	if !ok {
		return nil
	}

	if ref, ok := m["$ref"].(string); ok {
		m = resolveRef(ref, spec)
		if m == nil {
			return nil
		}
	}

	param := &RawParam{}
	param.Name, _ = m["name"].(string)
	param.In, _ = m["in"].(string)
	param.Required, _ = m["required"].(bool)
	if param.In == "path" {
		param.Required = true
	}

	if schema, ok := toStringMap(m["schema"]); ok {
		param.Schema = schema
		param.Example = schema["example"]
		param.Type, _ = schema["type"].(string)
	} else {
		param.Type, _ = m["type"].(string)
		param.Example = m["example"]
	}

	return param
}

func parseRequestBody(rb any, spec *Spec) *RawBody {
	m, ok := toStringMap(rb)
	if !ok {
		return nil
	}

	if ref, ok := m["$ref"].(string); ok {
		m = resolveRef(ref, spec)
		if m == nil {
			return nil
		}
	}

	body := &RawBody{
		Content: make(map[string]RawMediaType),
	}
	body.Required, _ = m["required"].(bool)

	if content, ok := toStringMap(m["content"]); ok {
		for mediaType, mediaData := range content {
			if mediaMap, ok := toStringMap(mediaData); ok {
				mt := RawMediaType{}
				if schema, ok := toStringMap(mediaMap["schema"]); ok {
					mt.Schema = resolveSchema(schema, spec)
				}
				body.Content[mediaType] = mt
			}
		}
	}

	return body
}

func resolveRef(ref string, spec *Spec) map[string]any {
	if !strings.HasPrefix(ref, "#/") {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	var current any

	switch {
	case len(parts) >= 2 && parts[0] == "components" && spec.Components != nil:
		current = spec.Components
	case len(parts) >= 1 && parts[0] == "definitions" && spec.Definitions != nil:
		current = spec.Definitions
		parts = parts[1:]
		if m, ok := current.(map[string]any); ok {
			if len(parts) > 0 {
				current = m[parts[0]]
				parts = parts[1:]
			}
		}
		result, _ := toStringMap(current)
		return result
	default:
		return nil
	}

	for _, part := range parts[1:] {
		m, ok := toStringMap(current)
		if !ok {
			return nil
		}
		current = m[part]
	}

	result, _ := toStringMap(current)
	return result
}

func resolveSchema(schema map[string]any, spec *Spec) map[string]any {
	if ref, ok := schema["$ref"].(string); ok {
		resolved := resolveRef(ref, spec)
		if resolved != nil {
			return resolved
		}
	}
	return schema
}

func isHTTPMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
		return true
	}
	return false
}

func toStringMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m, ok := v.(map[any]any); ok {
		result := make(map[string]any, len(m))
		for k, val := range m {
			if ks, ok := k.(string); ok {
				result[ks] = val
			}
		}
		return result, true
	}
	return nil, false
}
