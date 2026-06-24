package email

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/model"
)

func NormalizeAttachments(cfg *config.Config, attachments []model.Attachment) ([]model.Attachment, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	maxCount := 5
	maxBytes := int64(10 * 1024 * 1024)
	allowed := allowedAttachmentTypes("pdf,doc,docx,xls,xlsx,csv,png,jpg,jpeg")
	if cfg != nil {
		if cfg.Email.MaxAttachments > 0 {
			maxCount = cfg.Email.MaxAttachments
		}
		if cfg.Email.MaxAttachmentBytes > 0 {
			maxBytes = cfg.Email.MaxAttachmentBytes
		}
		if strings.TrimSpace(cfg.Email.AllowedAttachmentTypes) != "" {
			allowed = allowedAttachmentTypes(cfg.Email.AllowedAttachmentTypes)
		}
	}

	if len(attachments) > maxCount {
		return nil, fmt.Errorf("too many attachments: max %d", maxCount)
	}

	out := make([]model.Attachment, 0, len(attachments))
	for i, attachment := range attachments {
		normalized, err := normalizeAttachment(attachment, allowed, maxBytes)
		if err != nil {
			return nil, fmt.Errorf("attachment %d: %w", i+1, err)
		}
		out = append(out, normalized)
	}

	return out, nil
}

func normalizeAttachment(att model.Attachment, allowed map[string]bool, maxBytes int64) (model.Attachment, error) {
	att.FileName = strings.TrimSpace(att.FileName)
	att.ContentType = strings.TrimSpace(att.ContentType)
	att.Path = strings.TrimSpace(att.Path)
	att.ContentID = strings.TrimSpace(att.ContentID)
	att.DataBase64 = strings.TrimSpace(att.DataBase64)

	if att.FileName == "" && att.Path != "" {
		att.FileName = filepath.Base(att.Path)
	}
	if att.FileName == "" {
		return model.Attachment{}, fmt.Errorf("file_name is required")
	}
	if att.Inline && att.ContentID == "" {
		return model.Attachment{}, fmt.Errorf("content_id is required for inline attachments")
	}
	if err := validateAllowedAttachmentType(att, allowed); err != nil {
		return model.Attachment{}, err
	}

	hasPath := att.Path != ""
	hasData := len(att.Data) > 0 || att.DataBase64 != ""
	if hasPath == hasData {
		return model.Attachment{}, fmt.Errorf("provide exactly one of path or data_base64")
	}

	if hasPath {
		if err := validateAttachmentPath(att.Path); err != nil {
			return model.Attachment{}, err
		}
		info, err := os.Stat(att.Path)
		if err != nil {
			return model.Attachment{}, fmt.Errorf("path %q: %w", att.Path, err)
		}
		if info.IsDir() {
			return model.Attachment{}, fmt.Errorf("path %q is a directory", att.Path)
		}
		if info.Size() > maxBytes {
			return model.Attachment{}, fmt.Errorf("file exceeds max attachment size")
		}
		return att, nil
	}

	if len(att.Data) == 0 {
		decoded, err := base64.StdEncoding.DecodeString(att.DataBase64)
		if err != nil {
			return model.Attachment{}, fmt.Errorf("invalid data_base64: %w", err)
		}
		att.Data = decoded
	}
	if int64(len(att.Data)) > maxBytes {
		return model.Attachment{}, fmt.Errorf("file exceeds max attachment size")
	}
	if att.DataBase64 == "" {
		att.DataBase64 = base64.StdEncoding.EncodeToString(att.Data)
	}

	return att, nil
}

func validateAttachmentPath(path string) error {
	clean := filepath.Clean(path)
	for _, part := range strings.Split(clean, string(os.PathSeparator)) {
		if part == ".." {
			return fmt.Errorf("path traversal is not allowed")
		}
	}
	return nil
}

func validateAllowedAttachmentType(att model.Attachment, allowed map[string]bool) error {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(att.FileName)), ".")
	if ext == "" && att.Path != "" {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(att.Path)), ".")
	}
	if ext == "" {
		return fmt.Errorf("attachment extension is required")
	}
	if !allowed[ext] {
		return fmt.Errorf("attachment type %q is not allowed", ext)
	}
	return nil
}

func allowedAttachmentTypes(raw string) map[string]bool {
	out := map[string]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(item)), ".")
		if item != "" {
			out[item] = true
		}
	}
	return out
}
