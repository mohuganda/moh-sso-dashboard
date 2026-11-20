package report_scheduler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
			actor, ok := modulehttp.RequestAccess(r)
			return scheduler.HTTPActor{
				UserID:        actor.Subject,
				HealthContext: scheduler.HealthContext{District: actor.HealthScope.District, Facility: actor.HealthScope.Facility},
				HasSystem:     actor.HasSystem,
				HasPermission: actor.HasPermission,
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
