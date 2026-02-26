package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
)

type DocumentHandler struct {
	documentService *service.DocumentService
	auditService    *service.AuditService
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
		ContentType:      doc.ContentType.String,
		SizeBytes:        doc.SizeBytes,
		ChecksumSHA256:   &doc.ChecksumSha256.String,
		StorageLocation:  doc.StorageLocationID,
		ObjectKey:        doc.ObjectKey,
		UploadedBy:       doc.UploadedBy,
		CreatedAt:        doc.CreatedAt.Time.Format(time.RFC3339),
	}
}

func NewDocumentHandler(
	documentService *service.DocumentService,
	auditService *service.AuditService,
) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		auditService:    auditService,
	}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {

	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "INVALID_USER", "invalid user id")
		return
	}

	// Parse form
	storageLocationStr := c.PostForm("storage_location")
	storageLocation, err := uuid.Parse(storageLocationStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE_LOCATION", "invalid storage location")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "FILE_REQUIRED", "file is required")
		return
	}
	defer file.Close()

	// Generate object key
	objectKey := uuid.New().String()

	// Compute checksum
	checksum, err := computeSHA256(file)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "CHECKSUM_FAILED", err.Error())
		return
	}

	input := service.CreateDocumentInput{
		OriginalFilename: header.Filename,
		ContentType:      header.Header.Get("Content-Type"),
		SizeBytes:        header.Size,
		ChecksumSHA256:   &checksum,
		StorageLocation:  storageLocation,
		ObjectKey:        objectKey,
		UploadedBy:       userID,
	}

	doc, err := h.documentService.CreateDocument(c.Request.Context(), input)
	if err != nil {
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

	response.OK(c, http.StatusOK, doc)
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

func (h *DocumentHandler) UpdateDocument(c *gin.Context)        {}
func (h *DocumentHandler) DownloadDocument(c *gin.Context)      {}
func (h *DocumentHandler) ListDocuments(c *gin.Context)         {}
func (h *DocumentHandler) ListDocumentProcesses(c *gin.Context) {}

func computeSHA256(r io.Reader) (string, error) {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}

	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum), nil
}
