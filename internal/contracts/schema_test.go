package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestOutputSchemasRequireAllDeclaredObjectProperties(t *testing.T) {
	for _, name := range []string{
		"router_result.schema.json",
		"lookup_request.schema.json",
		"lookup_response.schema.json",
	} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "schemas", name))
			if err != nil {
				t.Fatalf("read schema: %v", err)
			}
			var schema map[string]any
			if err := json.Unmarshal(data, &schema); err != nil {
				t.Fatalf("parse schema: %v", err)
			}
			var problems []string
			checkRequiredProperties(name, schema, &problems)
			if len(problems) > 0 {
				t.Fatalf("schema is not strict enough for Codex structured output:\n%s", strings.Join(problems, "\n"))
			}
		})
	}
}

func checkRequiredProperties(path string, schema map[string]any, problems *[]string) {
	properties, ok := schema["properties"].(map[string]any)
	if ok {
		required := requiredSet(schema["required"])
		var missing []string
		for name := range properties {
			if !required[name] {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			*problems = append(*problems, path+" missing required properties: "+strings.Join(missing, ", "))
		}
		for name, raw := range properties {
			child, ok := raw.(map[string]any)
			if ok {
				checkRequiredProperties(path+".properties."+name, child, problems)
			}
		}
	}

	items, ok := schema["items"].(map[string]any)
	if ok {
		checkRequiredProperties(path+".items", items, problems)
	}
}

func requiredSet(raw any) map[string]bool {
	set := map[string]bool{}
	values, ok := raw.([]any)
	if !ok {
		return set
	}
	for _, value := range values {
		name, ok := value.(string)
		if ok {
			set[name] = true
		}
	}
	return set
}
