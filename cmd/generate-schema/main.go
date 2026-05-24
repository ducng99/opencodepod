package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"opencodepod/internal/config"
)

func main() {
	schema := generateSchema(reflect.TypeOf(config.Config{}))
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["title"] = "Config"
	schema["description"] = "OpenCodePod server configuration schema."

	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("config.schema.json", out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Schema written to config.schema.json")
}

func generateSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		items := generateSchema(t.Elem())
		return map[string]any{
			"type":  "array",
			"items": items,
		}
	case reflect.Map:
		additional := generateSchema(t.Elem())
		return map[string]any{
			"type":                 "object",
			"additionalProperties": additional,
		}
	case reflect.Struct:
		properties := map[string]any{}
		required := []string{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			name := jsonTag
			if len(name) > 0 {
				// strip options like omitempty
				for j := 0; j < len(jsonTag); j++ {
					if jsonTag[j] == ',' {
						name = jsonTag[:j]
						break
					}
				}
			}

			fieldSchema := generateSchema(field.Type)
			if desc := field.Tag.Get("desc"); desc != "" {
				fieldSchema["description"] = desc
			}
			properties[name] = fieldSchema

			if field.Tag.Get("required") == "true" {
				required = append(required, name)
			}
		}

		result := map[string]any{
			"type":       "object",
			"properties": properties,
		}
		if len(required) > 0 {
			result["required"] = required
		}
		return result
	default:
		return map[string]any{}
	}
}
