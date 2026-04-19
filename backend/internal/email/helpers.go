package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/model"
)

func hasBody(msg model.Message) bool {
	return strings.TrimSpace(msg.TextBody) != "" ||
		strings.TrimSpace(msg.HTMLBody) != "" ||
		strings.TrimSpace(msg.TemplateName) != ""
}


func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("marshal json: %w", err))
	}
	return b
}

func newReadSeeker(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}
