package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-api/configs"
	"go-api/models"

	"github.com/gin-gonic/gin"
)

// CreateReport creates a new report
func CreateReport(c *gin.Context) {
	var report models.Report

	if err := c.BindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// If setting as default, ensure only one default per thematic area
	if report.IsDefault && report.ThematicID > 0 {
		_, err := configs.DB.Exec(
			"UPDATE dwhportal.reports SET \"isDefault\"=false WHERE \"thematicId\"=$1",
			report.ThematicID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update defaults"})
			return
		}
	}

	// Insert into database
	_, err := configs.DB.Exec(
		"INSERT INTO dwhportal.reports(title, period, type, format, reportname, \"isDefault\", \"thematicId\") VALUES($1, $2, $3, $4, $5, $6, $7)",
		report.Title, report.Period, report.Type, report.Format, report.Reportname, report.IsDefault, report.ThematicID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "report created"})
}

// GetReport gets a single report by ID
func GetReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var report models.Report

	err := configs.DB.QueryRow(
		`SELECT id, title, period, type, format, reportname, "isDefault", "thematicId", "createdAt", "updatedAt" FROM dwhportal.reports WHERE id=$1`,
		id,
	).Scan(&report.ID, &report.Title, &report.Period, &report.Type, &report.Format, &report.Reportname, &report.IsDefault, &report.ThematicID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"report": report,
	})
}

// ListReports lists reports with pagination and optional thematicId filter
func ListReports(c *gin.Context) {
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
			`SELECT COUNT(*) FROM dwhportal.reports WHERE "thematicId"=$1`,
			thematicIDInt,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		// Fetch data with thematic join
		rows, err = configs.DB.Query(
			`SELECT r.id, r.title, r.period, r.type, r.format, r.reportname, r."isDefault", r."thematicId", r."createdAt", r."updatedAt",
				t.id, t."name", t.description, t.icon, t."createdAt", t."updatedAt"
			FROM dwhportal.reports r
			LEFT JOIN dwhportal.thematics t ON r."thematicId" = t.id
			WHERE r."thematicId"=$1
			ORDER BY r."createdAt" DESC
			LIMIT $2 OFFSET $3`,
			thematicIDInt, limit, offset,
		)
	} else {
		// Get total count
		err = configs.DB.QueryRow(
			`SELECT COUNT(*) FROM dwhportal.reports`,
		).Scan(&totalRecords)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count failed"})
			return
		}

		// Fetch data with thematic join
		rows, err = configs.DB.Query(
			`SELECT r.id, r.title, r.period, r.type, r.format, r.reportname, r."isDefault", r."thematicId", r."createdAt", r."updatedAt",
				t.id, t."name", t.description, t.icon, t."createdAt", t."updatedAt"
			FROM dwhportal.reports r
			LEFT JOIN dwhportal.thematics t ON r."thematicId" = t.id
			ORDER BY r."createdAt" DESC
			LIMIT $1 OFFSET $2`,
			limit, offset,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed", "details": err.Error()})
		return
	}
	defer rows.Close()

	// Calculate total pages
	totalPages := (totalRecords + limit - 1) / limit

	var reports []models.ReportWithThematic
	for rows.Next() {
		var r models.Report
		var t models.Thematic
		var reportWithThematic models.ReportWithThematic

		err := rows.Scan(
			&r.ID, &r.Title, &r.Period, &r.Type, &r.Format, &r.Reportname, &r.IsDefault, &r.ThematicID, &r.CreatedAt, &r.UpdatedAt,
			&t.ID, &t.Name, &t.Description, &t.Icon, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			continue
		}

		reportWithThematic.Report = r
		reportWithThematic.Thematic = &t
		reports = append(reports, reportWithThematic)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"results":      len(reports),
		"totalRecords": totalRecords,
		"totalPages":   totalPages,
		"currentPage":  page,
		"report":       reports,
	})
}

// GetReportsByTheme gets all reports for a specific thematic area
func GetReportsByTheme(c *gin.Context) {
	thematicID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := configs.DB.Query(
		`SELECT id, title, period, type, format, reportname, "isDefault", "thematicId", "createdAt", "updatedAt" FROM dwhportal.reports WHERE "thematicId"=$1 ORDER BY "createdAt" DESC`,
		thematicID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch failed", "details": err.Error()})
		return
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var r models.Report
		err := rows.Scan(&r.ID, &r.Title, &r.Period, &r.Type, &r.Format, &r.Reportname, &r.IsDefault, &r.ThematicID, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			continue
		}
		reports = append(reports, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"report": reports,
	})
}

// GetDefaultReport gets the default report for a thematic area
func GetDefaultReport(c *gin.Context) {
	thematicID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var report models.Report

	err := configs.DB.QueryRow(
		`SELECT id, title, period, type, format, reportname, "isDefault", "thematicId", "createdAt", "updatedAt" FROM dwhportal.reports WHERE "thematicId"=$1 AND "isDefault"=true`,
		thematicID,
	).Scan(&report.ID, &report.Title, &report.Period, &report.Type, &report.Format, &report.Reportname, &report.IsDefault, &report.ThematicID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "default report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"report": report,
	})
}

// UpdateReport updates a report
func UpdateReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var report models.Report
	if err := c.BindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// If setting as default, ensure only one default per thematic area
	if report.IsDefault {
		var currentThematicID int64
		err := configs.DB.QueryRow(
			`SELECT "thematicId" FROM dwhportal.reports WHERE id=$1`,
			id,
		).Scan(&currentThematicID)

		if err == nil {
			// Use provided thematicId or current one
			themeToUse := report.ThematicID
			if themeToUse == 0 {
				themeToUse = currentThematicID
			}

			// Clear other defaults in the same thematic area
			_, err = configs.DB.Exec(
				`UPDATE dwhportal.reports SET "isDefault"=false WHERE "thematicId"=$1 AND id!=$2`,
				themeToUse, id,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update defaults"})
				return
			}
		}
	}

	// Update in database
	res, err := configs.DB.Exec(
		`UPDATE dwhportal.reports SET title=$1, period=$2, type=$3, format=$4, reportname=$5, "isDefault"=$6, "thematicId"=$7, "updatedAt"=NOW() WHERE id=$8`,
		report.Title, report.Period, report.Type, report.Format, report.Reportname, report.IsDefault, report.ThematicID, id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	// Fetch updated record
	var updatedReport models.Report
	err = configs.DB.QueryRow(
		`SELECT id, title, period, type, format, reportname, "isDefault", "thematicId", "createdAt", "updatedAt" FROM dwhportal.reports WHERE id=$1`,
		id,
	).Scan(&updatedReport.ID, &updatedReport.Title, &updatedReport.Period, &updatedReport.Type, &updatedReport.Format, &updatedReport.Reportname, &updatedReport.IsDefault, &updatedReport.ThematicID, &updatedReport.CreatedAt, &updatedReport.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "updated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"report": updatedReport,
	})
}

// DeleteReport deletes a report
func DeleteReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := configs.DB.Exec(
		"DELETE FROM dwhportal.reports WHERE id=$1",
		id,
	)

	affected, _ := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
