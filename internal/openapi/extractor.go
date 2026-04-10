package openapi

import (
	"fmt"
	"strings"
)

type Operation struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Tags        []string
	PathParams  []Param
	QueryParams []Param
	HeaderParams []Param
	Body        map[string]any
	ContentType string
}

type Param struct {
	Name     string
	Required bool
	Example  any
	Type     string
}

func Extract(spec *Spec) []Operation {
	var ops []Operation

	for path, item := range spec.Paths {
		for method, rawOp := range item.Operations {
			op := Operation{
				Method:      method,
				Path:        path,
				OperationID: rawOp.OperationID,
				Summary:     rawOp.Summary,
				Tags:        rawOp.Tags,
			}

			for _, p := range rawOp.Parameters {
				param := Param{
					Name:     p.Name,
					Required: p.Required,
					Example:  p.Example,
					Type:     p.Type,
				}
				switch p.In {
				case "path":
					op.PathParams = append(op.PathParams, param)
				case "query":
					op.QueryParams = append(op.QueryParams, param)
				case "header":
					op.HeaderParams = append(op.HeaderParams, param)
				}
			}

			if rawOp.RequestBody != nil {
				op.Body, op.ContentType = buildBodyExample(rawOp.RequestBody, rawOp.Consumes, spec)
			}

			ops = append(ops, op)
		}
	}

	return ops
}

func buildBodyExample(rb *RawBody, consumes []string, spec *Spec) (map[string]any, string) {
	preferredTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, mt := range preferredTypes {
		if media, ok := rb.Content[mt]; ok {
			example := schemaToExample(media.Schema, spec, 0)
			if m, ok := example.(map[string]any); ok {
				return m, mt
			}
			return nil, mt
		}
	}

	if len(consumes) > 0 {
		return nil, consumes[0]
	}

	return nil, "application/json"
}

func schemaToExample(schema map[string]any, spec *Spec, depth int) any {
	if depth > 5 {
		return nil
	}

	if schema == nil {
		return nil
	}

	if ref, ok := schema["$ref"].(string); ok {
		resolved := resolveRef(ref, spec)
		if resolved != nil {
			return schemaToExample(resolved, spec, depth+1)
		}
		return nil
	}

	if example, ok := schema["example"]; ok {
		return example
	}

	schemaType, _ := schema["type"].(string)

	if schemaType == "object" || (schemaType == "" && schema["properties"] != nil) {
		return buildObjectExample(schema, spec, depth)
	}

	if schemaType == "array" {
		if items, ok := toStringMap(schema["items"]); ok {
			elem := schemaToExample(items, spec, depth+1)
			return []any{elem}
		}
		return []any{}
	}

	return primitiveExample(schemaType, schema)
}

func buildObjectExample(schema map[string]any, spec *Spec, depth int) map[string]any {
	result := make(map[string]any)

	props, ok := toStringMap(schema["properties"])
	if !ok {
		return result
	}

	for name, propData := range props {
		propSchema, ok := toStringMap(propData)
		if !ok {
			continue
		}
		result[name] = schemaToExample(propSchema, spec, depth+1)
	}

	return result
}

func primitiveExample(schemaType string, schema map[string]any) any {
	format, _ := schema["format"].(string)

	switch schemaType {
	case "string":
		switch format {
		case "email":
			return "user@example.com"
		case "date":
			return "2024-01-01"
		case "date-time":
			return "2024-01-01T00:00:00Z"
		case "uuid":
			return "00000000-0000-0000-0000-000000000000"
		case "password":
			return "secret"
		default:
			if enum, ok := schema["enum"].([]any); ok && len(enum) > 0 {
				return enum[0]
			}
			return "string"
		}
	case "integer", "number":
		return 1
	case "boolean":
		return true
	default:
		return nil
	}
}

func ResolvePath(path string, params []Param) string {
	result := path
	for _, p := range params {
		placeholder := fmt.Sprintf("{%s}", p.Name)
		value := paramValueString(p)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func paramValueString(p Param) string {
	if p.Example != nil {
		return fmt.Sprintf("%v", p.Example)
	}
	switch p.Type {
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	default:
		return p.Name
	}
}
