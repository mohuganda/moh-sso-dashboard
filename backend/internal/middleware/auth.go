package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/features/authsession"
	"github.com/moh-sso-dashboard/internal/keycloak"
)

func ExtractAuthContext(kc *keycloak.Client, sessions *authsession.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractAccessToken(c, sessions)
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

		authContext := authz.NewContext(user.ID, user.RealmRoles, user.ClientRoles)

		c.Set("access_token", accessToken)
		c.Set(authz.ContextKey, authContext)
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("client_roles", authContext.ClientRoles)
		c.Set("realm_roles", authContext.RealmRoles)
		c.Set("permissions", authContext.Permissions)
		c.Set("systems", authContext.AccessibleSystems)
		c.Set("is_admin", authContext.IsAdmin)
		c.Set("is_user", authContext.IsUser)

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
		if authContext, ok := authz.FromGin(c); ok && authContext.IsAdmin {
			c.Next()
			return
		}

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

func RequirePermission(permission authz.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		authContext, ok := authz.FromGin(c)
		if !ok || !authContext.HasPermission(permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":               "permission required",
				"required_permission": string(permission),
			})
			return
		}

		c.Next()
	}
}

func RequireAnyPermission(permissions ...authz.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		authContext, ok := authz.FromGin(c)
		if !ok || !authContext.HasAnyPermission(permissions...) {
			required := make([]string, 0, len(permissions))
			for _, permission := range permissions {
				required = append(required, string(permission))
			}

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":                "permission required",
				"required_permissions": required,
			})
			return
		}

		c.Next()
	}
}

func RequireRealmRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authContext, ok := authz.FromGin(c)
		if !ok || !authContext.HasRealmRole(role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "role required",
				"required_role": role,
			})
			return
		}

		c.Next()
	}
}

func RequireClientRole(clientIDParam string, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authContext, ok := authz.FromGin(c)
		clientID := strings.TrimSpace(c.Param(clientIDParam))
		if !ok || !authContext.HasClientRole(clientID, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":                "client role required",
				"required_client_role": role,
			})
			return
		}

		c.Next()
	}
}

func extractAccessToken(c *gin.Context, sessions *authsession.Store) string {
	// 1. Prefer Authorization header for API/mobile/backward compatibility.
	if token := extractBearerToken(c); token != "" {
		return token
	}

	// 2. Server-side session referenced by the opaque session cookie.
	if token := extractSessionToken(c, sessions); token != "" {
		return token
	}

	// 3. Legacy HttpOnly access_token cookie (pre-session deployments).
	if token := extractAccessTokenCookie(c); token != "" {
		return token
	}

	return ""
}

func extractSessionToken(c *gin.Context, sessions *authsession.Store) string {
	if sessions == nil {
		return ""
	}

	sessionID, err := c.Cookie(authsession.CookieName)
	if err != nil || sessionID == "" {
		return ""
	}

	sess, err := sessions.Get(c.Request.Context(), sessionID)
	if err != nil {
		return ""
	}

	return sess.AccessToken
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
