package routes

import (
	"github.com/gin-gonic/gin"

	authfeature "github.com/moh-sso-dashboard/internal/features/auth"
)

func RegisterAuthRoutes(api *gin.RouterGroup, deps Dependencies) {
	authfeature.RegisterRoutes(
		api,
		deps.Auth,
		deps.Limiter,
		deps.AuthLoginRateLimitPerMin,
		deps.AuthCallbackRateLimitPerMin,
		deps.AuthSessionRateLimitPerMinute,
	)
}
