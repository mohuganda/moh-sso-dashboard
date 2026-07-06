package geojson

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	basePath string
}

func NewHandler(basePath string) *Handler {
	return &Handler{
		basePath: basePath,
	}
}

func (h *Handler) GetGeoJSON(c *gin.Context) {
	name := c.Param("name")

	files := map[string]string{
		"geo_water":       "uga_water.geojson",
		"geo_districts":   "uganda_districts.geojson",
		"geo_subcounties": "uganda_subcounties.geojson",
	}

	filename, exists := files[name]
	if !exists {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "requested geojson does not exist")
		return
	}

	fullPath := filepath.Join(h.basePath, filename)

	file, err := os.Open(fullPath)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.IsDir() {
		response.Fail(c, http.StatusInternalServerError, "GEOJSON_LOAD_FAILED", "failed to load geojson file")
		return
	}

	c.Header("Content-Type", "application/geo+json")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, filename, stat.ModTime(), file)
}
