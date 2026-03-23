package handler

import (
	"go-api/configs"
	"go-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateThematic creates a new thematic area
func CreateThematic(c *gin.Context) {
	var thematic models.Thematic

	if err := c.BindJSON(&thematic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Insert into database (using dwhportal schema)
	_, err := configs.DB.Exec(
		"INSERT INTO dwhportal.thematics(\"name\", description, icon) VALUES($1, $2, $3)",
		thematic.Name, thematic.Description, thematic.Icon,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create thematic area"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "thematic area created"})
}

// GetThematic gets a single thematic area by ID
func GetThematic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var thematic models.Thematic

	err := configs.DB.QueryRow(
		"SELECT id, \"name\", description, icon, \"createdAt\", \"updatedAt\" FROM dwhportal.thematics WHERE id=$1",
		id,
	).Scan(&thematic.ID, &thematic.Name, &thematic.Description, &thematic.Icon, &thematic.CreatedAt, &thematic.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "thematic area not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"theme":  thematic,
	})
}

// ListThematics lists thematic areas with pagination
func ListThematics(c *gin.Context) {
	// Read query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Get total count
	var totalRecords int
	err := configs.DB.QueryRow(
		"SELECT COUNT(*) FROM dwhportal.thematics",
	).Scan(&totalRecords)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed", "details": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (totalRecords + limit - 1) / limit

	// Fetch data
	rows, err := configs.DB.Query(
		"SELECT id, \"name\", description, icon, \"createdAt\", \"updatedAt\" FROM dwhportal.thematics ORDER BY id DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed", "details": err.Error()})
		return
	}
	defer rows.Close()

	var thematics []models.Thematic
	for rows.Next() {
		var t models.Thematic
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		thematics = append(thematics, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"results":      len(thematics),
		"totalRecords": totalRecords,
		"totalPages":   totalPages,
		"currentPage":  page,
		"thematic":     thematics,
	})
}

// GetAllThematics gets all thematic areas without pagination
func GetAllThematics(c *gin.Context) {
	rows, err := configs.DB.Query(
		"SELECT id, \"name\", description, icon, \"createdAt\", \"updatedAt\" FROM dwhportal.thematics ORDER BY id DESC",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed", "details": err.Error()})
		return
	}
	defer rows.Close()

	var thematics []models.Thematic
	for rows.Next() {
		var t models.Thematic
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Icon, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		thematics = append(thematics, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"thematic": thematics,
	})
}

// UpdateThematic updates a thematic area
func UpdateThematic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var thematic models.Thematic
	if err := c.BindJSON(&thematic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Update in database
	res, err := configs.DB.Exec(
		"UPDATE dwhportal.thematics SET \"name\"=$1, description=$2, icon=$3, \"updatedAt\"=NOW() WHERE id=$4",
		thematic.Name, thematic.Description, thematic.Icon, id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "thematic area not found"})
		return
	}

	// Fetch updated record
	var updatedThematic models.Thematic
	err = configs.DB.QueryRow(
		"SELECT id, \"name\", description, icon, \"createdAt\", \"updatedAt\" FROM dwhportal.thematics WHERE id=$1",
		id,
	).Scan(&updatedThematic.ID, &updatedThematic.Name, &updatedThematic.Description, &updatedThematic.Icon, &updatedThematic.CreatedAt, &updatedThematic.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "updated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"thematic": updatedThematic,
	})
}

// DeleteThematic deletes a thematic area
func DeleteThematic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := configs.DB.Exec(
		"DELETE FROM dwhportal.thematics WHERE id=$1",
		id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "thematic area not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetThematicCount gets count of dashboards and uploads per thematic area
func GetThematicCount(c *gin.Context) {
	query := `
		SELECT 
			combined.id,
			combined.name,
			COUNT(*) AS count
		FROM (
			SELECT CAST(t.id AS INTEGER) AS id, t.name
			FROM dwhportal.dashboards d
			INNER JOIN dwhportal.thematics t ON CAST(t.id AS INTEGER) = CAST(d."thematicId" AS INTEGER)
			UNION ALL
			SELECT CAST(t.id AS INTEGER) AS id, t."name"
			FROM uploads u
			INNER JOIN dwhportal.thematics t ON CAST(t.id AS INTEGER) = CAST(u."thematicId" AS INTEGER)
		) AS combined
		GROUP BY combined.id, combined.name
	`

	rows, err := configs.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type ThematicCount struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var results []ThematicCount
	for rows.Next() {
		var tc ThematicCount
		err := rows.Scan(&tc.ID, &tc.Name, &tc.Count)
		if err != nil {
			continue
		}
		results = append(results, tc)
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No Category Details Found."})
		return
	}

	c.JSON(http.StatusOK, results)
}
