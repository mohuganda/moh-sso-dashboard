package routes

import (
	"github.com/gin-gonic/gin"

	surveillancefeature "github.com/moh-sso-dashboard/internal/features/surveillance"
)

func registerSurveillanceRoutes(protected *gin.RouterGroup, deps Dependencies) {
	surveillancefeature.RegisterProtectedRoutes(protected, deps.Surveillance)
}
