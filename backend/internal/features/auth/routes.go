package auth

import (
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
			ratelimit.MiddlewareForPolicy(limiter, ratelimit.LoginPolicy(loginRateLimitPerMin)),
			handler.HandleAuthLogin,
		)
		auth.GET(
			"/callback",
			ratelimit.MiddlewareForPolicy(
				limiter,
				ratelimit.PerIPPolicy(ratelimit.PolicyAuthCallback, callbackRateLimitPerMin),
			),
			handler.HandleAuthCallback,
		)
		auth.GET("/me", handler.HandleAuthGetMe)
		auth.POST(
			"/refresh",
			ratelimit.MiddlewareForPolicy(
				limiter,
				ratelimit.PerIPPolicy(ratelimit.PolicyAuthRefresh, sessionRateLimitPerMin),
			),
			handler.HandleAuthRefreshToken,
		)
		auth.GET("/logout", handler.HandleAuthLogout)
	}
}
