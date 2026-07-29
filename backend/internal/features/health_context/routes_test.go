package health_context

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterAdminRoutesSharesExistingUserWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	admin := router.Group("/api/v1/admin")
	admin.GET("/users/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	RegisterAdminRoutes(admin, &Handler{})
}
