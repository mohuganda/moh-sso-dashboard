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
	return isCSV(contentType, filename)
}
