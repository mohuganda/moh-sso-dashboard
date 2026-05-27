package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/storage"
)

type DocumentHandler struct {
	documentService        *service.DocumentService
	auditService           *service.AuditService
	storageLocationService service.StorageLocationService
	storage                storage.Storage
	storageFactory         *storage.StorageFactory
}

type UpdateDocumentRequest struct {
	OriginalFilename *string `json:"original_filename"`
	ContentType      *string `json:"content_type"`
}

type DocumentResponse struct {
	ID               uuid.UUID `json:"id"`
	OriginalFilename string    `json:"original_filename"`
	ContentType      string    `json:"content_type"`
	SizeBytes        int64     `json:"size_bytes"`
	ChecksumSHA256   *string   `json:"checksum_sha256,omitempty"`
	StorageLocation  uuid.UUID `json:"storage_location"`
	ObjectKey        string    `json:"object_key"`
	UploadedBy       uuid.UUID `json:"uploaded_by"`
	Status           string    `json:"status"`
	ObjectURL        string    `json:"object_url,omitempty"`
	ViewURL          string    `json:"view_url,omitempty"`
	DownloadURL      string    `json:"download_url,omitempty"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

func toDocumentResponse(doc db.Document, objectURL, viewURL, downloadURL string) DocumentResponse {
	return DocumentResponse{
		ID:               doc.ID,
		OriginalFilename: doc.OriginalFilename,
		ContentType:      nullStringValue(doc.ContentType),
		SizeBytes:        doc.SizeBytes,
		ChecksumSHA256:   nullStringPtr(doc.ChecksumSha256),
		StorageLocation:  doc.StorageLocationID,
		ObjectKey:        doc.ObjectKey,
		UploadedBy:       doc.UploadedBy,
		Status:           normalizeStatus(doc.Status),
		ObjectURL:        objectURL,
		ViewURL:          viewURL,
		DownloadURL:      downloadURL,
		CreatedAt:        doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        doc.UpdatedAt.Format(time.RFC3339),
	}
}

type ProcessResponse struct {
	ID          uuid.UUID  `json:"id"`
	DocumentID  uuid.UUID  `json:"document_id"`
	ProcessType string     `json:"process_type"`
	Status      string     `json:"status"`
	Progress    int32      `json:"progress"`
	Message     *string    `json:"message,omitempty"`
	Error       *string    `json:"error,omitempty"`
	Attempts    int32      `json:"attempts"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func toProcessResponse(process db.Process) ProcessResponse {
	return ProcessResponse{
		ID:          process.ID,
		DocumentID:  process.DocumentID,
		ProcessType: process.ProcessType,
		Status:      normalizeStatus(process.Status),
		Progress:    process.Progress,
		Message:     nullStringPtr(process.Message),
		Error:       nullStringPtr(process.Error),
		Attempts:    process.Attempts,
		CreatedBy:   process.CreatedBy,
		StartedAt:   nullTimePtr(process.StartedAt),
		FinishedAt:  nullTimePtr(process.FinishedAt),
		CreatedAt:   nullTimePtr(process.CreatedAt),
		UpdatedAt:   nullTimePtr(process.UpdatedAt),
	}
}

func NewDocumentHandler(
	documentService *service.DocumentService,
	auditService *service.AuditService,
	storageLocationService service.StorageLocationService,
	storage storage.Storage,
	storageFactory *storage.StorageFactory,
) *DocumentHandler {
	return &DocumentHandler{
		documentService:        documentService,
		auditService:           auditService,
		storageLocationService: storageLocationService,
		storage:                storage,
		storageFactory:         storageFactory,
	}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	ctx := c.Request.Context()

	log.Printf("[CreateDocument] Request received")

	userIDStr := c.GetString("user_id")

	log.Printf("[CreateDocument] user_id=%s", userIDStr)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[CreateDocument] invalid user id: %v", err)

		response.Fail(c, http.StatusUnauthorized, "INVALID_USER", "invalid user id")
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("[CreateDocument] multipart parse failed: %v", err)

		response.Fail(c, http.StatusBadRequest, "INVALID_MULTIPART", "invalid multipart form")
		return
	}

	storageLocationStr := c.PostForm("storage_location")

	log.Printf("[CreateDocument] storage_location=%s", storageLocationStr)

	storageLocationID, err := uuid.Parse(storageLocationStr)
	if err != nil {
		log.Printf("[CreateDocument] invalid storage location: %v", err)

		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE_LOCATION", "invalid storage location")
		return
	}

	isTemplate := strings.EqualFold(
		strings.TrimSpace(c.PostForm("is_template")),
		"true",
	)

	log.Printf("[CreateDocument] is_template=%v", isTemplate)

	metadata := map[string]any{}

	metadataRaw := strings.TrimSpace(c.PostForm("metadata"))

	log.Printf("[CreateDocument] metadata_raw=%s", metadataRaw)

	if metadataRaw != "" {
		if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {

			log.Printf("[CreateDocument] metadata unmarshal failed: %v", err)

			response.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_METADATA",
				"metadata must be valid JSON",
			)
			return
		}
	}

	log.Printf("[CreateDocument] parsed metadata=%v", metadata)

	templateCode := ""
	if value, ok := metadata["template_code"].(string); ok {
		templateCode = strings.TrimSpace(value)
	}

	log.Printf("[CreateDocument] template_code=%s", templateCode)

	file, header, err := c.Request.FormFile("file")
	if err != nil {

		log.Printf("[CreateDocument] file missing: %v", err)

		response.Fail(c, http.StatusBadRequest, "FILE_REQUIRED", "file is required")
		return
	}
	defer file.Close()

	log.Printf(
		"[CreateDocument] file received | filename=%s | size=%d",
		header.Filename,
		header.Size,
	)

	if header.Size == 0 {

		log.Printf("[CreateDocument] empty file uploaded")

		response.Fail(c, http.StatusBadRequest, "EMPTY_FILE", "file is empty")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	log.Printf("[CreateDocument] content_type=%s", contentType)

	fileNeedsProcessing := requiresProcessing(contentType, header.Filename)

	log.Printf(
		"[CreateDocument] needs_processing=%v",
		fileNeedsProcessing,
	)

	var processType model.ProcessType

	if fileNeedsProcessing && !isTemplate {

		processTypeValue := strings.TrimSpace(c.PostForm("process_type"))

		log.Printf("[CreateDocument] process_type_raw=%s", processTypeValue)

		if processTypeValue == "" {

			log.Printf("[CreateDocument] process type missing")

			response.Fail(
				c,
				http.StatusBadRequest,
				"PROCESS_TYPE_REQUIRED",
				"process type is required for this file type",
			)
			return
		}

		parsedProcessType, ok := model.ParseProcessType(processTypeValue)
		if !ok {

			log.Printf("[CreateDocument] invalid process type=%s", processTypeValue)

			response.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_PROCESS_TYPE",
				"invalid process type",
			)
			return
		}

		processType = parsedProcessType

		log.Printf("[CreateDocument] parsed process type=%s", processType)

		if templateCode == "" {

			log.Printf("[CreateDocument] template code missing")

			response.Fail(
				c,
				http.StatusBadRequest,
				"TEMPLATE_CODE_REQUIRED",
				"template code is required for processable document uploads",
			)
			return
		}
	}

	loc, err := h.storageLocationService.GetByID(ctx, storageLocationStr)
	if err != nil {

		log.Printf("[CreateDocument] storage lookup failed: %v", err)

		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE", err.Error())
		return
	}

	log.Printf(
		"[CreateDocument] storage location resolved | provider=%s",
		loc.Provider,
	)

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {

		log.Printf("[CreateDocument] storage provider init failed: %v", err)

		response.Fail(c, http.StatusBadRequest, "INVALID_PROVIDER", err.Error())
		return
	}

	objectKey := uuid.New().String()

	log.Printf("[CreateDocument] generated object_key=%s", objectKey)

	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	if err := storageProvider.Upload(
		ctx,
		objectKey,
		teeReader,
		header.Size,
		contentType,
	); err != nil {

		log.Printf("[CreateDocument] upload failed: %v", err)

		response.Fail(
			c,
			http.StatusInternalServerError,
			"UPLOAD_FAILED",
			"failed to upload file",
		)
		return
	}

	log.Printf("[CreateDocument] upload successful")

	checksumStr := hex.EncodeToString(hasher.Sum(nil))

	log.Printf("[CreateDocument] checksum=%s", checksumStr)

	status := db.DocumentStatusCOMPLETED
	if fileNeedsProcessing && !isTemplate {
		status = db.DocumentStatusPENDING
	}

	log.Printf("[CreateDocument] resolved status=%s", status)

	if templateCode != "" {
		metadata["template_code"] = templateCode
	}

	log.Printf("[CreateDocument] final metadata=%v", metadata)

	input := service.CreateDocumentInput{
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		SizeBytes:        header.Size,
		ChecksumSHA256:   &checksumStr,
		StorageLocation:  storageLocationID,
		ObjectKey:        objectKey,
		UploadedBy:       userID,
		Status:           status,
		IsTemplate:       isTemplate,
		Metadata:         metadata,
	}

	if fileNeedsProcessing && !isTemplate {
		input.ProcessType = processType
	}

	log.Printf("[CreateDocument] creating document service entry")

	doc, err := h.documentService.CreateDocument(ctx, input)
	if err != nil {

		log.Printf("[CreateDocument] document creation failed: %v", err)

		_ = storageProvider.Delete(ctx, objectKey)

		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_FAILED",
			err.Error(),
		)
		return
	}

	log.Printf(
		"[CreateDocument] document created successfully | document_id=%s",
		doc.ID,
	)

	objectURL, _ := storageProvider.GetObjectURL(ctx, objectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, objectKey, header.Filename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, objectKey, header.Filename)

	response.OK(
		c,
		http.StatusCreated,
		toDocumentResponse(doc, objectURL, viewURL, downloadURL),
	)
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", err.Error())
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", err.Error())
		return
	}

	objectURL, _ := storageProvider.GetObjectURL(ctx, doc.ObjectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, doc.ObjectKey, doc.OriginalFilename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, doc.ObjectKey, doc.OriginalFilename)

	response.OK(c, http.StatusOK, toDocumentResponse(doc, objectURL, viewURL, downloadURL))
}

func (h *DocumentHandler) EditDocument(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.OriginalFilename == nil || req.ContentType == nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "original_filename and content_type are required")
		return
	}

	doc, err := h.documentService.EditDocument(ctx, service.EditDocumentInput{
		ID:               id,
		OriginalFilename: *req.OriginalFilename,
		ContentType:      *req.ContentType,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", err.Error())
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", err.Error())
		return
	}

	objectURL, _ := storageProvider.GetObjectURL(ctx, doc.ObjectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, doc.ObjectKey, doc.OriginalFilename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, doc.ObjectKey, doc.OriginalFilename)

	response.OK(c, http.StatusOK, toDocumentResponse(doc, objectURL, viewURL, downloadURL))
}

func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	if err := h.documentService.DeleteDocument(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, gin.H{
		"message": "document deleted successfully",
	})
}

func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid uuid")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, uid)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", err.Error())
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", err.Error())
		return
	}

	reader, err := storageProvider.Download(ctx, doc.ObjectKey)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DOWNLOAD_FAILED", "failed to retrieve file")
		return
	}
	defer reader.Close()

	contentType := nullStringValue(doc.ContentType)
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.OriginalFilename))
	c.Header("Content-Type", contentType)

	_, _ = io.Copy(c.Writer, reader)
}

func (h *DocumentHandler) ViewDocument(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid uuid")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, uid)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", err.Error())
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", err.Error())
		return
	}

	reader, err := storageProvider.Download(ctx, doc.ObjectKey)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "VIEW_FAILED", "failed to retrieve file")
		return
	}
	defer reader.Close()

	contentType := nullStringValue(doc.ContentType)
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, doc.OriginalFilename))
	c.Header("Content-Type", contentType)

	_, _ = io.Copy(c.Writer, reader)
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	page := model.Pagination{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	docs, err := h.documentService.ListDocuments(ctx, page)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "failed to list documents")
		return
	}

	out := make([]DocumentResponse, 0, len(docs))
	for _, doc := range docs {
		objectURL := ""
		viewURL := ""
		downloadURL := ""

		loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
		if err == nil {
			storageProvider, err := h.storageFactory.Get(loc.Provider)
			if err == nil {
				objectURL, _ = storageProvider.GetObjectURL(ctx, doc.ObjectKey)
				viewURL, _ = storageProvider.GetViewURL(ctx, doc.ObjectKey, doc.OriginalFilename)
				downloadURL, _ = storageProvider.GetDownloadURL(ctx, doc.ObjectKey, doc.OriginalFilename)
			}
		}

		out = append(out, toDocumentResponse(doc, objectURL, viewURL, downloadURL))
	}

	response.OK(c, http.StatusOK, out)
}

func (h *DocumentHandler) ListDocumentProcesses(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "document id is required")
		return
	}

	processes, err := h.documentService.ListProcessesByDocument(ctx, idParam)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "PROCESS_LIST_FAILED", "failed to list processes")
		return
	}

	out := make([]ProcessResponse, 0, len(processes))
	for _, p := range processes {
		out = append(out, toProcessResponse(p))
	}

	response.OK(c, http.StatusOK, out)
}

func (h *DocumentHandler) ReprocessDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "error", "invalid document id")
		return
	}

	err = h.documentService.Reprocess(c.Request.Context(), documentID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "error", err.Error())
		return
	}

	response.OK(c, http.StatusOK, gin.H{"message": "reprocessing started"})
}

func normalizeStatus(v any) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	case []byte:
		return string(s)
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprint(v)
	}
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func nullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func nullTimeRFC3339(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.UTC().Format(time.RFC3339)
}

func requiresProcessing(mimeType, fileName string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "text/csv",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	case "application/pdf":
		return false
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".csv", ".xls", ".xlsx":
		return true
	case ".pdf":
		return false
	default:
		return false
	}
}
