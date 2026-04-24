package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type GeoJSONHandler struct {
	basePath string
}

func NewGeoJSONHandler(basePath string) *GeoJSONHandler {
	return &GeoJSONHandler{
		basePath: basePath,
	}
}

func (h *GeoJSONHandler) GetGeoJSON(c *gin.Context) {
	name := c.Param("name")

	files := map[string]string{
		"geo_water":       "uga_water.geojson",
		"geo_districts":   "uganda_districts.geojson",
		"geo_subcounties": "uganda_subcounties.geojson",
	}

	filename, exists := files[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "requested geojson does not exist",
		})
		return
	}

	fullPath := filepath.Join(h.basePath, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to load geojson file",
			"error":   err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/geo+json")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "application/geo+json", data)
}
