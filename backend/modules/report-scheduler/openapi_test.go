package reportscheduler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAPIMatchesHTTPRoutes(t *testing.T) {
	module, err := NewHTTPModule(NewService(nil, nil, nil, nil, nil, nil), HTTPOptions{
		Actor:               func(*http.Request) (HTTPActor, bool) { return HTTPActor{}, false },
		LimitPublicDownload: func(_ string, _ int, next http.Handler) http.Handler { return next },
	})
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]struct {
			Security    json.RawMessage `json:"security"`
			System      string          `json:"x-required-system"`
			Permissions [][]string      `json:"x-required-permission-groups"`
		}
	}
	if err := json.Unmarshal(module.OpenAPI(), &document); err != nil {
		t.Fatal(err)
	}
	public := map[string]bool{}
	for _, pattern := range module.PublicPatterns() {
		public[pattern] = true
	}
	count := 0
	for _, operations := range document.Paths {
		count += len(operations)
	}
	if count != len(module.Patterns()) {
		t.Fatalf("documented %d operations for %d routes", count, len(module.Patterns()))
	}
	for _, pattern := range module.Patterns() {
		method, path, _ := strings.Cut(pattern, " ")
		op, ok := document.Paths[path][strings.ToLower(method)]
		if !ok {
			t.Fatalf("undocumented route %s", pattern)
		}
		if public[pattern] {
			if string(op.Security) != "[]" || op.System != "" || len(op.Permissions) != 0 {
				t.Errorf("public route %s has unexpected security metadata: %+v", pattern, op)
			}
		} else if len(op.Security) != 0 || op.System != RequiredSystem || len(op.Permissions) == 0 {
			t.Errorf("protected route %s has unexpected security metadata: %+v", pattern, op)
		}
	}
}
