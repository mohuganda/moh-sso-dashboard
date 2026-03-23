package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// GetRProxy proxies requests to the R server for report generation
func GetRProxy(c *gin.Context) {
	// Get query parameters
	district := c.Query("district")
	year := c.Query("year")
	period := c.Query("period")
	reportName := c.Query("report_name")
	format := c.Query("format")

	// Validate required parameters
	if district == "" || year == "" || period == "" || reportName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing required query parameters: district, year, period, or reportname",
		})
		return
	}

	// Determine endpoint based on format
	endpoint := "reports"
	if format == "html" {
		endpoint = "render_html"
	}

	// Build R server URL
	rServerURL := fmt.Sprintf(
		"http://172.27.1.94:8080/%s?district=%s&period=%s&year=%s&report_name=%s",
		endpoint,
		url.QueryEscape(district),
		url.QueryEscape(period),
		url.QueryEscape(year),
		url.QueryEscape(reportName),
	)

	// Create HTTP client with long timeout (10 minutes)
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	// Create request
	req, err := http.NewRequest("GET", rServerURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error creating request",
			"error":   err.Error(),
		})
		return
	}

	// Set headers
	if format == "html" {
		req.Header.Set("Accept", "text/html")
	} else {
		req.Header.Set("Accept", "application/pdf")
	}
	req.Header.Set("Connection", "keep-alive")

	// Make request to R server
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching R report",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"message": "R server returned an error",
			"error":   fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
		return
	}

	// Set response headers
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	if format == "html" {
		// Stream HTML response
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Cache-Control", "no-cache")
		c.DataFromReader(http.StatusOK, resp.ContentLength, "text/html; charset=utf-8", resp.Body, nil)
	} else {
		// Send PDF response
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s_%s_%s.pdf", reportName, district, period))

		// Read and send PDF data
		pdfData, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error reading PDF data",
				"error":   err.Error(),
			})
			return
		}

		c.Data(http.StatusOK, "application/pdf", pdfData)
	}
}
