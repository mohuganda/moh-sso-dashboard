package geojson

import "github.com/gin-gonic/gin"

func RegisterProtectedRoutes(protected *gin.RouterGroup, handler *Handler) {
	geojson := protected.Group("/geojson")
	{
		geojson.GET("/:name", handler.GetGeoJSON)
	}
}
