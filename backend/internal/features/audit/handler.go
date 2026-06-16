package audit

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	"github.com/moh-sso-dashboard/internal/cache"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
)

/* =========================================================
 * Handler
 * ========================================================= */

type Handler struct {
	store db.Store
	cache *cache.RedisCache
}

func NewHandler(store db.Store, cache *cache.RedisCache) *Handler {
	return &Handler{store: store,
		cache: cache}
}

/* =========================================================
 * Helpers
 * ========================================================= */

func mustParseTimeRFC3339(c *gin.Context, key string) (time.Time, bool) {
	v := c.Query(key)
	if v == "" {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", key+" is required (RFC3339)")
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_DATE_FORMAT", "Invalid "+key+" format (RFC3339)")
		return time.Time{}, false
	}
	return t, true
}

func parseUUIDParam(v string) (*uuid.UUID, bool) {
	if v == "" {
		return nil, true
	}
	id, err := uuid.Parse(v)
	if err != nil {
		return nil, false
	}
	return &id, true
}

/* =========================================================
 * List Audit Logs (cursor pagination)
 * ========================================================= */

func (h *Handler) ListAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	cacheKey := auditLogsCacheKey(c)

	if h.cache != nil {
		var cached gin.H
		if ok, _ := h.cache.Get(c.Request.Context(), cacheKey, &cached); ok {
			response.OK(c, http.StatusOK, cached)
			return
		}
	}

	// filters
	action := c.Query("action")
	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success")

	var userID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		id, ok := parseUUIDParam(v)
		if !ok {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user_id format")
			return
		}
		userID = id
	}

	// cursor
	var cursorCreatedAt *time.Time
	var cursorID *uuid.UUID

	if v := c.Query("cursor_created_at"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_DATE_FORMAT", "Invalid cursor_created_at")
			return
		}
		cursorCreatedAt = &t
	}

	if v := c.Query("cursor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid cursor_id")
			return
		}
		cursorID = &id
	}

	limit := int32(50)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "limit must be between 1 and 200")
			return
		}
		limit = int32(n)
	}

	rows, err := h.store.ListAuditLogs(
		c.Request.Context(),
		db.ListAuditLogsParams{
			StartTime: from,
			EndTime:   to,

			Action:   toNullString(action),
			UserID:   toNullUUID(userID),
			ClientID: toNullString(clientID),
			Ip:       toNullString(ip),
			Success:  toNullString(success),

			CursorCreatedAt: toNullTime(cursorCreatedAt),
			CursorID:        toNullUUID(cursorID),

			RowLimit: limit + 1,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit logs")
		return
	}

	hasMore := len(rows) > int(limit)
	if hasMore {
		rows = rows[:limit]
	}

	var nextCreatedAt *time.Time
	var nextID *uuid.UUID

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		if last.CreatedAt.Valid {
			t := last.CreatedAt.Time
			nextCreatedAt = &t
		}
		id := last.ID
		nextID = &id
	}

	resp := toAuditLogListResponse(
		toAuditLogResponses(rows),
		toAuditCursorResponse(nextCreatedAt, nextID),
		hasMore,
	)

	if h.cache != nil {
		_ = h.cache.Set(
			c.Request.Context(),
			cacheKey,
			resp,
			20*time.Second, // perfect for audit logs
		)
	}

	response.OK(c, http.StatusOK, resp)
}

/* =========================================================
 * Get Single Audit Log
 * ========================================================= */

func (h *Handler) GetAuditLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid audit log ID")
		return
	}

	row, err := h.store.GetAuditLog(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "AUDIT_LOG_NOT_FOUND", "Audit log not found")
		return
	}

	response.OK(c, http.StatusOK, toAuditLogResponseFromGet(row))
}

/* =========================================================
 * Audit Metadata / Metrics
 * ========================================================= */

func (h *Handler) ListAuditActions(c *gin.Context) {
	rows, err := h.store.ListAuditActions(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit actions")
		return
	}

	response.OK(c, http.StatusOK, toAuditActionsResponse(rows))
}

func (h *Handler) AuditMetricsOverview(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	row, err := h.store.AuditMetricsOverview(
		c.Request.Context(),
		db.AuditMetricsOverviewParams{
			StartTime: toNullTime(&from),
			EndTime:   toNullTime(&to),
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute audit metrics")
		return
	}

	response.OK(c, http.StatusOK, toAuditMetricsOverviewResponse(row))
}

func (h *Handler) FailedLoginsByDay(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	rows, err := h.store.FailedLoginsByDay(
		c.Request.Context(),
		db.FailedLoginsByDayParams{
			StartTime: toNullTime(&from),
			EndTime:   toNullTime(&to),
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute failed logins series")
		return
	}

	response.OK(c, http.StatusOK, toAuditFailedLoginsByDayResponse(rows))
}

func (h *Handler) TopFailureIPs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	limit := int32(10)
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "limit must be between 1 and 50")
			return
		}
		limit = int32(n)
	}

	rows, err := h.store.TopFailureIPs(
		c.Request.Context(),
		db.TopFailureIPsParams{
			StartTime: toNullTime(&from),
			EndTime:   toNullTime(&to),
			RowLimit:  limit,
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute top IPs")
		return
	}

	response.OK(c, http.StatusOK, toAuditTopFailureIPsResponse(rows))
}

/* =========================================================
 * Export (CSV / JSON) – streaming, no envelope
 * ========================================================= */

func (h *Handler) ExportAuditLogs(c *gin.Context) {
	from, ok := mustParseTimeRFC3339(c, "from")
	if !ok {
		return
	}
	to, ok := mustParseTimeRFC3339(c, "to")
	if !ok {
		return
	}

	format := c.DefaultQuery("format", "csv")
	action := c.Query("action")
	clientID := c.Query("client_id")
	ip := c.Query("ip")
	success := c.Query("success")

	var userID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		id, ok := parseUUIDParam(v)
		if !ok {
			response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user_id format")
			return
		}
		userID = id
	}

	rows, err := h.store.ExportAuditLogs(
		c.Request.Context(),
		db.ExportAuditLogsParams{
			StartTime: toNullTime(&from),
			EndTime:   toNullTime(&to),
			Action:    toNullString(action),
			UserID:    toNullUUID(userID),
			ClientID:  toNullString(clientID),
			Ip:        toNullString(ip),
			Success:   toNullString(success),
		},
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Export failed")
		return
	}

	items := toExportAuditLogResponses(rows)
	manifest := buildAuditManifest(from, to, items, len(items), c)

	if format == "json" {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=audit_logs_export.json")
		response.OK(c, http.StatusOK, toAuditExportResponse(manifest, items))
		return
	}

	writeAuditCSV(c, items, manifest)
}

/* =========================================================
 * Helpers (export + sql nulls)
 * ========================================================= */

func buildAuditManifest(from, to time.Time, rows any, recordCount int, c *gin.Context) AuditExportManifest {
	return toAuditExportManifest(from, to, rows, recordCount)
}

func writeAuditCSV(c *gin.Context, rows []AuditLogResponse, manifest AuditExportManifest) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs_export.csv")

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	_ = w.Write([]string{
		"created_at", "user_id", "username", "action",
		"ip", "client_id", "success", "metadata_json",
	})

	for _, r := range rows {
		metaJSON, _ := json.Marshal(r.Metadata)
		success := ""
		if r.Success != nil {
			success = strconv.FormatBool(*r.Success)
		}

		_ = w.Write([]string{
			r.CreatedAt,
			r.UserID,
			r.Username,
			r.Action,
			r.IP,
			r.ClientID,
			success,
			string(metaJSON),
		})
	}

	manifestBytes, _ := json.Marshal(manifest)
	c.Header("X-Audit-Export-Manifest", string(manifestBytes))
}

/* ---- null helpers ---- */

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func extractAuditMetadata(r pqtype.NullRawMessage) map[string]any {
	if !r.Valid || len(r.RawMessage) == 0 {
		return nil
	}

	var meta map[string]any
	if err := json.Unmarshal(r.RawMessage, &meta); err != nil {
		return nil
	}
	return meta
}

func auditLogsCacheKey(c *gin.Context) string {
	return fmt.Sprintf(
		"audit_logs:%s",
		c.Request.URL.RawQuery,
	)
}
