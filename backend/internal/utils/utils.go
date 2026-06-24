package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
)

func GenerateClientID() string {
	return uuid.New().String()
}

// helper
func ToNullUUID(s string) uuid.NullUUID {
	if s == "" {
		return uuid.NullUUID{Valid: false}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func ExtractUserIDFromJWT(token string) string {
	if token == "" {
		return ""
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return ""
	}

	// Keycloak stores user ID in "sub"
	if sub, ok := data["sub"].(string); ok {
		return sub
	}

	return ""
}

func Encode(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func ProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "app.env")); err == nil {
			return wd
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			log.Fatal("app.env not found in any parent directory")
		}
		wd = parent
	}
}

func NormalizeHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func SplitClientIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func ParseBoolDefaultTrue(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return true, nil
	}
	switch s {
	case "true", "1", "yes", "y":
		return true, nil
	case "false", "0", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid enabled value: %q", s)
	}
}

func EscapeCSV(s string) string {
	s = strings.ReplaceAll(s, `"`, `""`)
	if strings.ContainsAny(s, ",\n\r") {
		return `"` + s + `"`
	}
	return s
}

func TokenHasRealmRole(tokenStr string, role string) bool {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return false
	}

	claims := token.Claims.(jwt.MapClaims)

	ra, ok := claims["realm_access"].(map[string]interface{})
	if !ok {
		return false
	}

	roles, ok := ra["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, r := range roles {
		if r.(string) == role {
			return true
		}
	}

	return false
}

func MustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func DefaultClientAttributes(clientID string) map[string]string {
	sidenav := []map[string]string{
		{
			"id":         "dashboard",
			"label":      "Dashboard",
			"path":       "/dashboard",
			"permission": clientID + ".dashboard.view",
		},
		{
			"id":         "settings",
			"label":      "Settings",
			"path":       "/settings",
			"permission": clientID + ".settings.manage",
		},
	}

	raw, _ := json.Marshal(sidenav)

	return map[string]string{
		"ui.sidenav": string(raw),
		"ui.home":    "/dashboard",
		"ui.icon":    "applications",
	}
}

func ValidateRedirectURIs(uris []string) error {
	for _, uri := range uris {
		uri = strings.TrimSpace(uri)
		if uri == "" {
			return fmt.Errorf("redirect URI cannot be empty")
		}

		if uri == "*" {
			return fmt.Errorf("wildcard redirect URI '*' is not allowed")
		}

		parsed, err := url.Parse(uri)
		if err != nil {
			return fmt.Errorf("invalid redirect URI %q: %w", uri, err)
		}

		if parsed.Scheme == "" {
			return fmt.Errorf("redirect URI %q must include a scheme", uri)
		}

		if parsed.Scheme == "http" && parsed.Host != "localhost" {
			return fmt.Errorf(
				"insecure redirect URI %q: only localhost may use http",
				uri,
			)
		}

		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf(
				"unsupported redirect URI scheme %q",
				parsed.Scheme,
			)
		}
	}

	return nil
}

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

func NewReadSeeker(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}

func ValidateMessage(msg model.Message) error {
	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}

	for _, to := range msg.To {
		if strings.TrimSpace(to.Email) == "" {
			return errors.New("recipient email cannot be empty")
		}
	}

	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("subject is required")
	}

	if strings.TrimSpace(msg.TextBody) == "" {
		return errors.New("body is required")
	}

	return nil
}
