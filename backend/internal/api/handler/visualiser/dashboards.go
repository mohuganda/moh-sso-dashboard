package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-api/configs"
	"go-api/models"

	"github.com/gin-gonic/gin"
)

// CreateDashboard creates a new dashboard
func CreateDashboard(c *gin.Context) {
	var dashboard models.Dashboard

	if err := c.BindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Insert into database
	_, err := configs.DB.Exec(
		`INSERT INTO dwhportal.dashboards("name", description, url, image, "thematicId") VALUES($1, $2, $3, $4, $5)`,
		dashboard.Name, dashboard.Description, dashboard.URL, dashboard.Image, dashboard.ThematicID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dashboard"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "dashboard created"})
}

// GetDashboard gets a single dashboard by ID
func GetDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var dashboard models.Dashboard

	err := configs.DB.QueryRow(
		`SELECT id, "name", description, url, image, "thematicId", "createdAt", "updatedAt" FROM dwhportal.dashboards WHERE id=$1`,
		id,
	).Scan(&dashboard.ID, &dashboard.Name, &dashboard.Description, &dashboard.URL, &dashboard.Image, &dashboard.ThematicID, &dashboard.CreatedAt, &dashboard.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"dashboard": dashboard,
	})
}

// ListDashboards lists dashboards with pagination and optional thematicId filter
func ListDashboards(c *gin.Context) {
	// Read query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	thematicID := c.Query("thematicId")
	offset := (page - 1) * limit

	var totalRecords int
	var rows *sql.Rows
	var err error

	// Build query based on whether thematicId is provided
	if thematicID != "" {
		thematicIDInt, _ := strconv.ParseInt(thematicID, 10, 64)

		// Get total count
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.dashboards WHERE "thematicId"=$1`,
			thematicIDInt,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		// Fetch data
		rows, err = configs.DB.Query(
			`SELECT id, "name", description, url, image, "thematicId", "createdAt", "updatedAt" FROM dwhportal.dashboards WHERE "thematicId"=$1 ORDER BY "createdAt" DESC LIMIT $2 OFFSET $3`,
			thematicIDInt, limit, offset,
		)
	} else {
		// Get total count
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.dashboards`,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		// Fetch data
		rows, err = configs.DB.Query(
			`SELECT id, "name", description, url, image, "thematicId", "createdAt", "updatedAt" FROM dwhportal.dashboards ORDER BY "createdAt" DESC LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}
	defer rows.Close()

	// Calculate total pages
	totalPages := (totalRecords + limit - 1) / limit

	var dashboards []models.Dashboard
	for rows.Next() {
		var d models.Dashboard
		err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.URL, &d.Image, &d.ThematicID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			continue
		}
		dashboards = append(dashboards, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"results":      len(dashboards),
		"totalRecords": totalRecords,
		"totalPages":   totalPages,
		"currentPage":  page,
		"dashboard":    dashboards,
	})
}

// GetDashboardsByTheme gets all dashboards for a specific thematic area
func GetDashboardsByTheme(c *gin.Context) {
	thematicID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := configs.DB.Query(
		`SELECT id, "name", description, url, image, "thematicId", "createdAt", "updatedAt" FROM dwhportal.dashboards WHERE "thematicId"=$1 ORDER BY "createdAt" DESC`,
		thematicID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed"})
		return
	}
	defer rows.Close()

	var dashboards []models.Dashboard
	for rows.Next() {
		var d models.Dashboard
		err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.URL, &d.Image, &d.ThematicID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			continue
		}
		dashboards = append(dashboards, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"dashboard": dashboards,
	})
}

// UpdateDashboard updates a dashboard
func UpdateDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var dashboard models.Dashboard
	if err := c.BindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Update in database
	res, err := configs.DB.Exec(
		`UPDATE dwhportal.dashboards SET "name"=$1, description=$2, url=$3, image=$4, "thematicId"=$5, "updatedAt"=NOW() WHERE id=$6`,
		dashboard.Name, dashboard.Description, dashboard.URL, dashboard.Image, dashboard.ThematicID, id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	// Fetch updated record
	var updatedDashboard models.Dashboard
	err = configs.DB.QueryRow(
		`SELECT id, "name", description, url, image, "thematicId", "createdAt", "updatedAt" FROM dwhportal.dashboards WHERE id=$1`,
		id,
	).Scan(&updatedDashboard.ID, &updatedDashboard.Name, &updatedDashboard.Description, &updatedDashboard.URL, &updatedDashboard.Image, &updatedDashboard.ThematicID, &updatedDashboard.CreatedAt, &updatedDashboard.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "updated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"dashboard": updatedDashboard,
	})
}

// DeleteDashboard deletes a dashboard
func DeleteDashboard(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := configs.DB.Exec(
		"DELETE FROM dwhportal.dashboards WHERE id=$1",
		id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
