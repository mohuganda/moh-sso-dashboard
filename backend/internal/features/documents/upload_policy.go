package documents

import "strings"

func shouldQueueProcessing(
	contentType string,
	filename string,
	isTemplate bool,
	templateCode string,
) bool {
	if isTemplate || !requiresProcessing(contentType, filename) {
		return false
	}
	if strings.TrimSpace(templateCode) != "" {
		return true
	}
	// No template selected: only CSV can still be queued, since it needs no
	// template-defined structure to parse (unlike Excel).
	return isCSV(contentType, filename)
}
