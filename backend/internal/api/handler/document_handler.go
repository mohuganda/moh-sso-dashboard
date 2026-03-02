package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	CreatedAt        string    `json:"created_at"`
}

func toDocumentResponse(doc db.Document) DocumentResponse {
	return DocumentResponse{
		ID:               doc.ID,
		OriginalFilename: doc.OriginalFilename,
		ContentType:      nullStringValue(doc.ContentType),
		SizeBytes:        doc.SizeBytes,
		ChecksumSHA256:   nullStringPtr(doc.ChecksumSha256),
		StorageLocation:  doc.StorageLocationID,
		ObjectKey:        doc.ObjectKey,
		UploadedBy:       doc.UploadedBy,
		CreatedAt:        nullTimeRFC3339(doc.CreatedAt),
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

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "INVALID_USER", "invalid user id")
		return
	}

	storageLocationStr := c.PostForm("storage_location")
	storageLocationID, err := uuid.Parse(storageLocationStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE_LOCATION", "invalid storage location")
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB
		response.Fail(c, http.StatusBadRequest, "INVALID_MULTIPART", "invalid multipart form")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "FILE_REQUIRED", "file is required")
		return
	}
	defer file.Close()

	if header.Size == 0 {
		response.Fail(c, http.StatusBadRequest, "EMPTY_FILE", "file is empty")
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, storageLocationStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE", err.Error())
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PROVIDER", err.Error())
		return
	}

	// 5️⃣ Upload while computing checksum (streaming)
	objectKey := uuid.New().String()

	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := storageProvider.Upload(
		ctx,
		objectKey,
		teeReader,
		header.Size,
		contentType,
	); err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload file")
		return
	}

	checksumStr := hex.EncodeToString(hasher.Sum(nil))

	input := service.CreateDocumentInput{
		OriginalFilename: header.Filename,
		ContentType:      contentType,
		SizeBytes:        header.Size,
		ChecksumSHA256:   &checksumStr,
		StorageLocation:  storageLocationID,
		ObjectKey:        objectKey,
		UploadedBy:       userID,
	}

	doc, err := h.documentService.CreateDocument(ctx, input)
	if err != nil {
		// 🔥 Rollback uploaded file if DB fails (same provider!)
		_ = storageProvider.Delete(ctx, objectKey)

		response.Fail(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusCreated, toDocumentResponse(doc))
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	doc, err := h.documentService.GetDocument(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	response.OK(c, http.StatusOK, toDocumentResponse(doc))
}

func (h *DocumentHandler) EditDocument(c *gin.Context) {
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

	doc, err := h.documentService.EditDocument(c.Request.Context(), service.EditDocumentInput{
		ID:               id,
		OriginalFilename: *req.OriginalFilename,
		ContentType:      *req.ContentType,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, doc)
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
		response.Fail(c, http.StatusInternalServerError, "INVALID_UUID", "Invalid Uuid")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, uid)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "Document not found")
		return
	}

	reader, err := h.storage.Download(ctx, doc.ObjectKey)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DOWNLOAD_FAILED", "Failed to retrieve file")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.OriginalFilename))
	c.Header("Content-Type", "application/octet-stream")

	io.Copy(c.Writer, reader)
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	model := model.Pagination{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	docs, err := h.documentService.ListDocuments(ctx, model)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list documents")
		return
	}

	response.OK(c, http.StatusOK, docs)
}

func (h *DocumentHandler) ListDocumentProcesses(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "INVALID_DOCUMENT_ID", "Document ID is required")
		return
	}

	processes, err := h.documentService.ListProcessesByDocument(ctx, idParam)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "PROCESS_LIST_FAILED", "Failed to list processes")
		return
	}

	out := make([]ProcessResponse, 0, len(processes))
	for _, p := range processes {
		out = append(out, toProcessResponse(p))
	}

	response.OK(c, http.StatusOK, out)
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
