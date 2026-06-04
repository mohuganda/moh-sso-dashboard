package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterAuthRoutes(api *gin.RouterGroup, deps Dependencies) {
	auth := api.Group("/auth")
	{
		auth.GET(
			"/login",
			ratelimit.Middleware(deps.Limiter, ratelimit.ByIP, deps.AuthLoginRateLimitPerMin, time.Minute),
			deps.Auth.HandleAuthLogin,
		)
		auth.GET(
			"/callback",
			ratelimit.Middleware(deps.Limiter, ratelimit.ByIP, deps.AuthCallbackRateLimitPerMin, time.Minute),
			deps.Auth.HandleAuthCallback,
		)
		auth.GET(
			"/me",
			ratelimit.Middleware(deps.Limiter, ratelimit.ByIP, deps.AuthSessionRateLimitPerMinute, time.Minute),
			deps.Auth.HandleAuthGetMe,
		)
		auth.POST(
			"/refresh",
			ratelimit.Middleware(deps.Limiter, ratelimit.ByIP, deps.AuthSessionRateLimitPerMinute, time.Minute),
			deps.Auth.HandleAuthRefreshToken,
		)
		auth.GET("/logout", deps.Auth.HandleAuthLogout)
	}
}
