package utils

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
)

// makeJWT builds an unsigned-but-well-formed JWT (header.payload.signature)
// the way Keycloak emits them: base64url without padding, RS256 header.
func makeJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	header, err := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-key-id",
	})
	if err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}

	enc := base64.RawURLEncoding
	return enc.EncodeToString(header) + "." +
		enc.EncodeToString(payload) + "." +
		enc.EncodeToString([]byte("fake-signature"))
}

// Realistic Keycloak access token claims (realm MoH, client dashboard-web).
func keycloakAccessTokenClaims(sub string) map[string]interface{} {
	claims := map[string]interface{}{
		"exp":          1765553397,
		"iat":          1765551597,
		"jti":          "f59af150-9c4c-4a13-9862-3bdce2587673",
		"iss":          "https://auth.health.go.ug/realms/MoH",
		"aud":          "account",
		"typ":          "Bearer",
		"azp":          "dashboard-web",
		"sid":          "27971a2b-9a40-435d-8d94-96584d973084",
		"scope":        "openid profile email",
		"realm_access": map[string]interface{}{"roles": []string{"user"}},
		"resource_access": map[string]interface{}{
			"dashboard-web": map[string]interface{}{"roles": []string{"viewer"}},
		},
		"preferred_username": "jdoe",
		"email":              "jdoe@health.go.ug",
	}
	if sub != "" {
		claims["sub"] = sub
	}
	return claims
}

func TestExtractUserIDFromJWT(t *testing.T) {
	const userID = "8f14e45f-ceea-467f-a1d4-91ab40429c5e"

	t.Run("keycloak access token with sub", func(t *testing.T) {
		token := makeJWT(t, keycloakAccessTokenClaims(userID))
		if got := ExtractUserIDFromJWT(token); got != userID {
			t.Errorf("got %q, want %q", got, userID)
		}
	})

	t.Run("lightweight access token without sub", func(t *testing.T) {
		token := makeJWT(t, keycloakAccessTokenClaims(""))
		if got := ExtractUserIDFromJWT(token); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		if got := ExtractUserIDFromJWT(""); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("opaque non-JWT token", func(t *testing.T) {
		if got := ExtractUserIDFromJWT("not-a-jwt-token"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("garbage payload segment", func(t *testing.T) {
		if got := ExtractUserIDFromJWT("aGVhZGVy.!!!notbase64!!!.c2ln"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("non-string sub claim", func(t *testing.T) {
		token := makeJWT(t, map[string]interface{}{"sub": 12345})
		if got := ExtractUserIDFromJWT(token); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestExtractUserIDFromTokens(t *testing.T) {
	const userID = "8f14e45f-ceea-467f-a1d4-91ab40429c5e"

	idToken := makeJWT(t, map[string]interface{}{
		"exp":                1765553397,
		"iat":                1765551597,
		"iss":                "https://auth.health.go.ug/realms/MoH",
		"aud":                "dashboard-web",
		"typ":                "ID",
		"azp":                "dashboard-web",
		"sub":                userID,
		"sid":                "27971a2b-9a40-435d-8d94-96584d973084",
		"preferred_username": "jdoe",
	})

	t.Run("falls back to id token when access token lacks sub", func(t *testing.T) {
		lightweight := makeJWT(t, keycloakAccessTokenClaims(""))
		if got := ExtractUserIDFromTokens(lightweight, idToken); got != userID {
			t.Errorf("got %q, want %q", got, userID)
		}
	})

	t.Run("prefers access token sub when present", func(t *testing.T) {
		const accessSub = "11111111-2222-3333-4444-555555555555"
		access := makeJWT(t, keycloakAccessTokenClaims(accessSub))
		if got := ExtractUserIDFromTokens(access, idToken); got != accessSub {
			t.Errorf("got %q, want %q", got, accessSub)
		}
	})

	t.Run("all tokens unusable", func(t *testing.T) {
		if got := ExtractUserIDFromTokens("", "opaque"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestJWTClaimNames(t *testing.T) {
	t.Run("returns sorted claim names", func(t *testing.T) {
		token := makeJWT(t, map[string]interface{}{
			"sub": "x", "iss": "y", "aud": "z",
		})
		want := []string{"aud", "iss", "sub"}
		if got := JWTClaimNames(token); !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("nil for invalid token", func(t *testing.T) {
		if got := JWTClaimNames("opaque"); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})
}
