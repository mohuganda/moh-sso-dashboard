package documents

import "strings"

func shouldQueueProcessing(
	contentType string,
	filename string,
	isTemplate bool,
	templateCode string,
) bool {
	return !isTemplate &&
		strings.TrimSpace(templateCode) != "" &&
		requiresProcessing(contentType, filename)
}
