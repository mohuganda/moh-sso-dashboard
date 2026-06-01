package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/keycloak"
)

func ExtractAuthContext(kc *keycloak.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractAccessToken(c)
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing token",
			})
			return
		}

		user, err := kc.Me(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user session",
			})
			return
		}

		c.Set("access_token", accessToken)
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("client_roles", user.ClientRoles)
		c.Set("realm_roles", user.RealmRoles)
		c.Set("is_admin", user.IsAdmin)
		c.Set("is_user", user.IsUser)

		c.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		admin, ok := isAdmin.(bool)
		if !ok || !admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		c.Next()
	}
}

func extractAccessToken(c *gin.Context) string {
	// 1. Prefer Authorization header for API/mobile/backward compatibility.
	if token := extractBearerToken(c); token != "" {
		return token
	}

	// 2. Fallback to HttpOnly cookie for browser session auth.
	if token := extractAccessTokenCookie(c); token != "" {
		return token
	}

	return ""
}

func extractBearerToken(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func extractAccessTokenCookie(c *gin.Context) string {
	token, err := c.Cookie("access_token")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(token)
}
