package documents

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	storagelocationfeature "github.com/moh-sso-dashboard/internal/features/storage_locations"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/storage"
)

type Handler struct {
	documentService        *Service
	auditService           *sharedservice.AuditService
	storageLocationService storagelocationfeature.Service
	storage                storage.Storage
	storageFactory         *storage.StorageFactory
}

func NewHandler(
	documentService *Service,
	auditService *sharedservice.AuditService,
	storageLocationService storagelocationfeature.Service,
	storage storage.Storage,
	storageFactory *storage.StorageFactory,
) *Handler {
	return &Handler{
		documentService:        documentService,
		auditService:           auditService,
		storageLocationService: storageLocationService,
		storage:                storage,
		storageFactory:         storageFactory,
	}
}

func (h *Handler) CreateDocument(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.GetString("user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "INVALID_USER", "invalid user id")
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_MULTIPART", "invalid multipart form")
		return
	}

	storageLocationStr := c.PostForm("storage_location")

	storageLocationID, err := uuid.Parse(storageLocationStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE_LOCATION", "invalid storage location")
		return
	}

	isTemplate := strings.EqualFold(
		strings.TrimSpace(c.PostForm("is_template")),
		"true",
	)

	metadata := map[string]any{}

	metadataRaw := strings.TrimSpace(c.PostForm("metadata"))

	if metadataRaw != "" {
		if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
			response.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_METADATA",
				"metadata must be valid JSON",
			)
			return
		}
	}

	templateCode := ""
	if value, ok := metadata["template_code"].(string); ok {
		templateCode = strings.TrimSpace(value)
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

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileNeedsProcessing := requiresProcessing(contentType, header.Filename)

	var processType model.ProcessType

	if fileNeedsProcessing && !isTemplate {

		processTypeValue := strings.TrimSpace(c.PostForm("process_type"))

		if processTypeValue == "" {
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
			response.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_PROCESS_TYPE",
				"invalid process type",
			)
			return
		}

		processType = parsedProcessType

		if templateCode == "" {
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
		response.Fail(c, http.StatusBadRequest, "INVALID_STORAGE", "request failed")
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PROVIDER", "request failed")
		return
	}

	objectKey := uuid.New().String()

	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	if err := storageProvider.Upload(
		ctx,
		objectKey,
		teeReader,
		header.Size,
		contentType,
	); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"UPLOAD_FAILED",
			"failed to upload file",
		)
		return
	}

	checksumStr := hex.EncodeToString(hasher.Sum(nil))

	status := DocumentStatusCompleted
	if fileNeedsProcessing && !isTemplate {
		status = DocumentStatusPending
	}

	if templateCode != "" {
		metadata["template_code"] = templateCode
	}

	input := CreateDocumentInput{
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

	doc, err := h.documentService.CreateDocument(ctx, input)
	if err != nil {
		_ = storageProvider.Delete(ctx, objectKey)

		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_FAILED",
			"request failed",
		)
		return
	}

	objectURL, _ := storageProvider.GetObjectURL(ctx, objectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, objectKey, header.Filename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, objectKey, header.Filename)

	response.OK(
		c,
		http.StatusCreated,
		toDocumentResponse(doc, objectURL, viewURL, downloadURL),
	)
}

func (h *Handler) GetDocument(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "request failed")
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", "request failed")
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", "request failed")
		return
	}

	objectURL, _ := storageProvider.GetObjectURL(ctx, doc.ObjectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, doc.ObjectKey, doc.OriginalFilename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, doc.ObjectKey, doc.OriginalFilename)

	response.OK(c, http.StatusOK, toDocumentResponse(doc, objectURL, viewURL, downloadURL))
}

func (h *Handler) EditDocument(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "request failed")
		return
	}

	if req.OriginalFilename == nil || req.ContentType == nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "original_filename and content_type are required")
		return
	}

	doc, err := h.documentService.EditDocument(ctx, EditDocumentInput{
		ID:               id,
		OriginalFilename: *req.OriginalFilename,
		ContentType:      *req.ContentType,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "request failed")
		return
	}

	loc, err := h.storageLocationService.GetByID(ctx, doc.StorageLocationID.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", "request failed")
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", "request failed")
		return
	}

	objectURL, _ := storageProvider.GetObjectURL(ctx, doc.ObjectKey)
	viewURL, _ := storageProvider.GetViewURL(ctx, doc.ObjectKey, doc.OriginalFilename)
	downloadURL, _ := storageProvider.GetDownloadURL(ctx, doc.ObjectKey, doc.OriginalFilename)

	response.OK(c, http.StatusOK, toDocumentResponse(doc, objectURL, viewURL, downloadURL))
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	if err := h.documentService.DeleteDocument(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DELETE_FAILED", "request failed")
		return
	}

	response.OK(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "document deleted successfully"})
}

func (h *Handler) DownloadDocument(c *gin.Context) {
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
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", "request failed")
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", "request failed")
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

func (h *Handler) ViewDocument(c *gin.Context) {
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
		response.Fail(c, http.StatusInternalServerError, "STORAGE_LOOKUP_FAILED", "request failed")
		return
	}

	storageProvider, err := h.storageFactory.Get(loc.Provider)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVALID_PROVIDER", "request failed")
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

func (h *Handler) ListDocuments(c *gin.Context) {
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

func (h *Handler) ListDocumentProcesses(c *gin.Context) {
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

func (h *Handler) ReprocessDocument(c *gin.Context) {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "error", "invalid document id")
		return
	}

	err = h.documentService.Reprocess(c.Request.Context(), documentID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "error", "request failed")
		return
	}

	response.OK(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "reprocessing started"})
}
