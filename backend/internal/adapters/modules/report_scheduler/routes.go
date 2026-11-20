package report_scheduler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/modulehttp"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	scheduler "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func RegisterProtectedRoutes(group *gin.RouterGroup, handler *Handler) {
	if handler == nil {
		return
	}
	module := newHTTPModule(handler)
	registrar := modulehttp.NewRegistrar(group)
	for _, pattern := range module.ProtectedPatterns() {
		registrar.Handle(pattern, module.ProtectedHandler())
	}
}

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler, limiter *ratelimit.Limiter) {
	if handler == nil {
		return
	}
	module := newHTTPModule(handler)
	registrar := modulehttp.NewRegistrar(group.Group("", ratelimit.MiddlewareForPolicy(limiter,
		ratelimit.PerIPPolicy(scheduler.DeliveryDownloadPolicy, scheduler.DeliveryDownloadLimit))))
	for _, pattern := range module.PublicPatterns() {
		registrar.Handle(pattern, module.PublicHandler())
	}
}

func newHTTPModule(handler *Handler) *scheduler.HTTPModule {
	module, err := scheduler.NewHTTPModule(handler.service, scheduler.HTTPOptions{
		Actor: func(r *http.Request) (scheduler.HTTPActor, bool) {
			actor, ok := modulehttp.Actor(r)
			health := modulehttp.HealthContext(r)
			return scheduler.HTTPActor{
				UserID:        actor.UserID,
				HealthContext: scheduler.HealthContext{District: strings.TrimSpace(health["district"]), Facility: strings.TrimSpace(health["facility"])},
				HasSystem:     actor.HasSystem,
				HasPermission: func(permission string) bool {
					return actor.HasPermission(authz.Permission(permission))
				},
			}, ok
		},
		RequestID: modulehttp.RequestID,
		// RegisterPublicRoutes installs the Gin limiter before dispatch, where
		// the dashboard's trusted-proxy client IP resolution is available.
		LimitPublicDownload: func(_ string, _ int, next http.Handler) http.Handler { return next },
	})
	if err != nil {
		panic(err)
	}
	return module
}
