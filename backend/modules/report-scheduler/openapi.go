package reportscheduler

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed openapi.json
var openAPIContract []byte

// OpenAPI describes the routes relative to the host API base path. Route
// authorization metadata comes from the same declarations used to mount them.
func (m *HTTPModule) OpenAPI() []byte {
	var document struct {
		Paths      map[string]map[string]map[string]any `json:"paths"`
		Components json.RawMessage                      `json:"components"`
	}
	if err := json.Unmarshal(openAPIContract, &document); err != nil {
		panic(err)
	}
	for _, route := range (&httpHandler{}).endpoints() {
		operation := document.Paths[route.fullPath()][strings.ToLower(route.method)]
		if operation == nil {
			panic("missing scheduler OpenAPI operation: " + route.pattern())
		}
		if route.public {
			operation["security"] = []any{}
		} else {
			delete(operation, "security")
			operation["x-required-system"] = RequiredSystem
			operation["x-required-permission-groups"] = route.required
		}
	}
	data, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	return data
}
