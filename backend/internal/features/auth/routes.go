package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/ratelimit"
)

func RegisterRoutes(
	api *gin.RouterGroup,
	handler *Handler,
	limiter *ratelimit.Limiter,
	loginRateLimitPerMin int,
	callbackRateLimitPerMin int,
	sessionRateLimitPerMin int,
) {
	auth := api.Group("/auth")
	{
		auth.GET(
			"/login",
			ratelimit.Middleware(limiter, ratelimit.ByIP, loginRateLimitPerMin, time.Minute),
			handler.HandleAuthLogin,
		)
		auth.GET(
			"/callback",
			ratelimit.Middleware(limiter, ratelimit.ByIP, callbackRateLimitPerMin, time.Minute),
			handler.HandleAuthCallback,
		)
		auth.GET(
			"/me",
			ratelimit.Middleware(limiter, ratelimit.ByIP, sessionRateLimitPerMin, time.Minute),
			handler.HandleAuthGetMe,
		)
		auth.POST(
			"/refresh",
			ratelimit.Middleware(limiter, ratelimit.ByIP, sessionRateLimitPerMin, time.Minute),
			handler.HandleAuthRefreshToken,
		)
		auth.GET("/logout", handler.HandleAuthLogout)
	}
}
