package apidocgen

import (
	"go/token"
	"go/types"
	"testing"
)

func TestRequiredPointerBooleanRejectsNullInContract(t *testing.T) {
	s := &schemas{definitions: object{}}
	field := types.NewVar(token.NoPos, nil, "Enabled", types.NewPointer(types.Typ[types.Bool]))
	for _, required := range []bool{false, true} {
		tag := `json:"enabled"`
		if required {
			tag += ` binding:"required"`
		}
		schema := s.of(types.NewStruct([]*types.Var{field}, []string{tag}))
		property := schema["properties"].(object)["enabled"].(object)
		if property["type"] != "boolean" {
			t.Fatalf("unexpected property: %v", property)
		}
		if required && property["nullable"] != nil {
			t.Fatalf("required pointer must reject null: %v", property)
		}
		if !required && property["nullable"] != true {
			t.Fatalf("optional pointer should allow null: %v", property)
		}
	}
}
