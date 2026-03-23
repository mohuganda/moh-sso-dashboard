package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-api/configs"
	"go-api/models"

	"github.com/gin-gonic/gin"
)

// CreateFeedback creates a new feedback (public route - no auth required)
func CreateFeedback(c *gin.Context) {
	var feedback models.Feedback

	if err := c.BindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Set default values if not provided
	if feedback.Status == "" {
		feedback.Status = "pending"
	}
	if feedback.Priority == "" {
		feedback.Priority = "medium"
	}

	// Insert into database
	_, err := configs.DB.Exec(
		`INSERT INTO dwhportal.feedbacks("name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes") VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		feedback.Name, feedback.Email, feedback.Message, feedback.Status, feedback.Priority,
		feedback.Screenshot, feedback.ReportID, feedback.ReportName, feedback.DashboardID, feedback.DashboardName, feedback.AdminNotes,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Thank you for your feedback!",
	})
}

// GetFeedback gets a single feedback by ID
func GetFeedback(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var feedback models.Feedback

	err := configs.DB.QueryRow(
		`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks WHERE id=$1`,
		id,
	).Scan(&feedback.ID, &feedback.Name, &feedback.Email, &feedback.Message, &feedback.Status, &feedback.Priority,
		&feedback.Screenshot, &feedback.ReportID, &feedback.ReportName, &feedback.DashboardID, &feedback.DashboardName,
		&feedback.AdminNotes, &feedback.CreatedAt, &feedback.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feedback not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"feedback": feedback,
	})
}

// ListFeedbacks lists feedbacks with pagination and optional filtering
func ListFeedbacks(c *gin.Context) {
	// Read query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	priority := c.Query("priority")
	offset := (page - 1) * limit

	var totalRecords int
	var rows *sql.Rows
	var err error

	// Build query based on filters
	if status != "" && priority != "" {
		// Both filters
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE status=$1 AND priority=$2`,
			status, priority,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		rows, err = configs.DB.Query(
			`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks WHERE status=$1 AND priority=$2 ORDER BY "createdAt" DESC LIMIT $3 OFFSET $4`,
			status, priority, limit, offset,
		)
	} else if status != "" {
		// Status filter only
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE status=$1`,
			status,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		rows, err = configs.DB.Query(
			`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks WHERE status=$1 ORDER BY "createdAt" DESC LIMIT $2 OFFSET $3`,
			status, limit, offset,
		)
	} else if priority != "" {
		// Priority filter only
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE priority=$1`,
			priority,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		rows, err = configs.DB.Query(
			`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks WHERE priority=$1 ORDER BY "createdAt" DESC LIMIT $2 OFFSET $3`,
			priority, limit, offset,
		)
	} else {
		// No filters
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.feedbacks`,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		rows, err = configs.DB.Query(
			`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks ORDER BY "createdAt" DESC LIMIT $1 OFFSET $2`,
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

	var feedbacks []models.Feedback
	for rows.Next() {
		var f models.Feedback
		err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Message, &f.Status, &f.Priority,
			&f.Screenshot, &f.ReportID, &f.ReportName, &f.DashboardID, &f.DashboardName,
			&f.AdminNotes, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			continue
		}
		feedbacks = append(feedbacks, f)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"results":      len(feedbacks),
		"totalRecords": totalRecords,
		"totalPages":   totalPages,
		"currentPage":  page,
		"feedback":     feedbacks,
	})
}

// UpdateFeedback updates a feedback (admin route - for status, priority, notes)
func UpdateFeedback(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var feedback models.Feedback
	if err := c.BindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Update in database
	res, err := configs.DB.Exec(
		`UPDATE dwhportal.feedbacks SET "name"=$1, email=$2, message=$3, status=$4, priority=$5, screenshot=$6, "reportId"=$7, "reportName"=$8, "dashboardId"=$9, "dashboardName"=$10, "adminNotes"=$11, "updatedAt"=NOW() WHERE id=$12`,
		feedback.Name, feedback.Email, feedback.Message, feedback.Status, feedback.Priority,
		feedback.Screenshot, feedback.ReportID, feedback.ReportName, feedback.DashboardID, feedback.DashboardName,
		feedback.AdminNotes, id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "feedback not found"})
		return
	}

	// Fetch updated record
	var updatedFeedback models.Feedback
	err = configs.DB.QueryRow(
		`SELECT id, "name", email, message, status, priority, screenshot, "reportId", "reportName", "dashboardId", "dashboardName", "adminNotes", "createdAt", "updatedAt" FROM dwhportal.feedbacks WHERE id=$1`,
		id,
	).Scan(&updatedFeedback.ID, &updatedFeedback.Name, &updatedFeedback.Email, &updatedFeedback.Message, &updatedFeedback.Status, &updatedFeedback.Priority,
		&updatedFeedback.Screenshot, &updatedFeedback.ReportID, &updatedFeedback.ReportName, &updatedFeedback.DashboardID, &updatedFeedback.DashboardName,
		&updatedFeedback.AdminNotes, &updatedFeedback.CreatedAt, &updatedFeedback.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Feedback updated successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Feedback updated successfully",
		"feedback": updatedFeedback,
	})
}

// DeleteFeedback deletes a feedback
func DeleteFeedback(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := configs.DB.Exec(
		"DELETE FROM dwhportal.feedbacks WHERE id=$1",
		id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "feedback not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetFeedbackStats gets feedback statistics (admin route)
func GetFeedbackStats(c *gin.Context) {
	// Get counts by status
	var total, pending, reviewed, resolved int

	err := configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks`).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stats failed"})
		return
	}

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE status='pending'`).Scan(&pending)
	if err != nil {
		pending = 0
	}

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE status='reviewed'`).Scan(&reviewed)
	if err != nil {
		reviewed = 0
	}

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE status='resolved'`).Scan(&resolved)
	if err != nil {
		resolved = 0
	}

	// Get counts by priority
	var low, medium, high int

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE priority='low'`).Scan(&low)
	if err != nil {
		low = 0
	}

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE priority='medium'`).Scan(&medium)
	if err != nil {
		medium = 0
	}

	err = configs.DB.QueryRow(`SELECT COUNT(*) FROM dwhportal.feedbacks WHERE priority='high'`).Scan(&high)
	if err != nil {
		high = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stats": gin.H{
			"total": total,
			"byStatus": gin.H{
				"pending":  pending,
				"reviewed": reviewed,
				"resolved": resolved,
			},
			"byPriority": gin.H{
				"low":    low,
				"medium": medium,
				"high":   high,
			},
		},
	})
}
