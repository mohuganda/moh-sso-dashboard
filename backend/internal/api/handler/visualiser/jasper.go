package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper function to create authenticated HTTP client
func createJasperClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Minute,
	}
}

// Helper function to create Basic Auth header
func getJasperAuth() string {
	username := os.Getenv("JASPER_USERNAME")
	password := os.Getenv("JASPER_PASSWORD")
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + credentials
}

// Helper function to get Jasper base URL
func getJasperBaseURL() string {
	baseURL := os.Getenv("JASPER_BASE_URL")
	// Remove trailing slashes
	baseURL = strings.TrimSuffix(baseURL, "/")
	return baseURL
}

// Helper function to forward response headers
func forwardHeaders(sourceHeaders http.Header, targetResponse gin.ResponseWriter) {
	for key, values := range sourceHeaders {
		if strings.ToLower(key) != "transfer-encoding" {
			for _, value := range values {
				targetResponse.Header().Add(key, value)
			}
		}
	}
}

// Helper function to rewrite URLs in HTML content
func rewriteJasperUrls(htmlContent string, jasperDomain string) string {
	// Escape special regex characters in domain
	escapedDomain := regexp.QuoteMeta(jasperDomain)

	// Determine proxy base URL
	proxyBaseURL := "http://localhost:9090/api"
	if os.Getenv("NODE_ENV") == "production" {
		proxyBaseURL = "https://dashboards.health.go.ug/api"
	}

	// Replace domain URLs
	domainRegex := regexp.MustCompile(fmt.Sprintf(`https?://%s`, escapedDomain))
	modifiedContent := domainRegex.ReplaceAllString(htmlContent, proxyBaseURL)

	// Replace relative jasperserver-pro URLs
	modifiedContent = regexp.MustCompile(`["']/jasperserver-pro/`).ReplaceAllString(modifiedContent, fmt.Sprintf(`"%s/jasperserver-pro/`, proxyBaseURL))

	return modifiedContent
}

// GetJasperReport gets a Jasper report and rewrites URLs
func GetJasperReport(c *gin.Context) {
	orgunit := c.Query("orgunit")
	period := c.Query("period")
	year := c.Query("year")
	reportname := c.Query("reportname")

	// Validate required parameters
	if orgunit == "" || period == "" || year == "" || reportname == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Missing required parameters: orgunit, period, year, reportname",
		})
		return
	}

	baseURL := getJasperBaseURL()
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Jasper server URL not configured",
		})
		return
	}

	// Construct Jasper report URL
	jasperURL := fmt.Sprintf(
		"%s/rest_v2/reports/Reports/%s.html?orgunit=%s&year=%s&period=%s",
		baseURL,
		reportname,
		orgunit,
		year,
		period,
	)

	// Create request
	req, err := http.NewRequest("GET", jasperURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error creating request",
			"error":   err.Error(),
		})
		return
	}

	// Set authentication header
	req.Header.Set("Authorization", getJasperAuth())

	// Make request
	client := createJasperClient()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch report from Jasper server",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode == 401 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication failed with Jasper server. Please check credentials.",
			"error":   "Unauthorized",
		})
		return
	}

	if resp.StatusCode == 404 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Report not found on Jasper server.",
			"error":   "Not Found",
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"message": "Failed to fetch report from Jasper server",
			"error":   fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
		return
	}

	// Read HTML content
	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error reading response from Jasper server",
			"error":   err.Error(),
		})
		return
	}

	// Rewrite URLs to use our proxy
	jasperDomain := ""
	if parsedURL, err := url.Parse(baseURL); err == nil {
		jasperDomain = parsedURL.Hostname()
	}

	var modifiedHTML string
	if jasperDomain != "" {
		modifiedHTML = rewriteJasperUrls(string(htmlContent), jasperDomain)
	} else {
		// Fallback: send original HTML without rewriting
		modifiedHTML = string(htmlContent)
		c.Header("X-URL-Rewrite-Failed", "true")
	}

	// Set response headers
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s_%s_P%s.html", reportname, orgunit, period))

	// Send HTML response
	c.String(http.StatusOK, modifiedHTML)
}

// ProxyJasperserverPro proxies requests to jasperserver-pro resources
func ProxyJasperserverPro(c *gin.Context) {
	baseURL := getJasperBaseURL()
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Jasper server URL not configured",
		})
		return
	}

	// Get path from wildcard parameter
	path := c.Param("path")
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	jasperURL := baseURL + "/" + path
	if c.Request.URL.RawQuery != "" {
		jasperURL += "?" + c.Request.URL.RawQuery
	}

	// Create request
	req, err := http.NewRequest("GET", jasperURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error creating request",
			"error":   err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", getJasperAuth())

	// Make request
	client := createJasperClient()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to proxy jasperserver-pro resource",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"message": "Failed to proxy jasperserver-pro resource",
			"error":   fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
		return
	}

	// Forward headers and stream response
	forwardHeaders(resp.Header, c.Writer)
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// ProxyRestV2 proxies requests to rest_v2 resources
func ProxyRestV2(c *gin.Context) {
	baseURL := getJasperBaseURL()
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Jasper server URL not configured",
		})
		return
	}

	// Get path from wildcard parameter
	path := c.Param("path")
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	jasperURL := baseURL + "/rest_v2/" + path
	if c.Request.URL.RawQuery != "" {
		jasperURL += "?" + c.Request.URL.RawQuery
	}

	// Create request
	req, err := http.NewRequest("GET", jasperURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error creating request",
			"error":   err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", getJasperAuth())

	// Make request
	client := createJasperClient()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to proxy rest_v2 resource",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"message": "Failed to proxy rest_v2 resource",
			"error":   fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
		return
	}

	// Forward headers and stream response
	forwardHeaders(resp.Header, c.Writer)
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// ProxyReportResource proxies requests to reportresource endpoint
func ProxyReportResource(c *gin.Context) {
	baseURL := getJasperBaseURL()
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Jasper server URL not configured",
		})
		return
	}

	jasperURL := baseURL + "/reportresource"
	if c.Request.URL.RawQuery != "" {
		jasperURL += "?" + c.Request.URL.RawQuery
	}

	// Create request
	req, err := http.NewRequest("GET", jasperURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error creating request",
			"error":   err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", getJasperAuth())

	// Make request
	client := createJasperClient()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to proxy reportresource",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"success": false,
			"message": "Failed to proxy reportresource",
			"error":   fmt.Sprintf("Status code: %d", resp.StatusCode),
		})
		return
	}

	// Forward headers and stream response
	forwardHeaders(resp.Header, c.Writer)
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
