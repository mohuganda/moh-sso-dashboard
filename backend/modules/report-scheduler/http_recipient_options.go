package reportscheduler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (h *httpHandler) RecipientOptions(c *httpRequest) {
	c.writer.Header().Set("Cache-Control", "private, no-store")
	kind := "email"
	if c.request.URL.Query().Has("kind") {
		kind = c.request.URL.Query().Get("kind")
	}
	if kind != "email" && kind != "group" && kind != "portal" {
		c.fail(400, "INVALID_KIND", "kind must be email, group, or portal")
		return
	}
	limit := 100
	if raw := c.request.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			c.fail(400, "INVALID_LIMIT", "limit must be between 1 and 100")
			return
		}
		limit = value
	}
	search := strings.TrimSpace(c.request.URL.Query().Get("search"))
	if len(search) > 200 {
		c.fail(400, "INVALID_SEARCH", "search is too long")
		return
	}
	if h.service == nil || !h.service.Module().SchedulingEnabled {
		c.fail(503, "DIRECTORY_UNAVAILABLE", "Recipient directory is unavailable")
		return
	}
	var selected []string
	if raw := c.request.URL.Query().Get("selected"); raw != "" {
		if len(raw) > 32768 || json.Unmarshal([]byte(raw), &selected) != nil || len(selected) > 100 {
			c.fail(400, "INVALID_SELECTION", "selected must contain at most 100 recipient values")
			return
		}
	}
	options, err := h.service.ListRecipientOptions(c.request.Context(), kind, search, limit, selected)
	if err != nil {
		c.fail(503, "DIRECTORY_UNAVAILABLE", "Unable to load recipient options")
		return
	}
	c.ok(http.StatusOK, options)
}
