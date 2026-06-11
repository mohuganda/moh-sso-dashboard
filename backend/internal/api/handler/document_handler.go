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
	db                     *sql.DB
	remoteDB               *sql.DB
}

type UpdateDocumentRequest struct {
	OriginalFilename *string `json:"original_filename"`
	ContentType      *string `json:"content_type"`
}

type DocumentResponse struct {
	ID               uuid.UUID              `json:"id"`
	OriginalFilename string                 `json:"original_filename"`
	ContentType      string                 `json:"content_type"`
	SizeBytes        int64                  `json:"size_bytes"`
	ChecksumSHA256   *string                `json:"checksum_sha256,omitempty"`
	StorageLocation  uuid.UUID              `json:"storage_location"`
	ObjectKey        string                 `json:"object_key"`
	UploadedBy       uuid.UUID              `json:"uploaded_by"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ObjectURL        string                 `json:"object_url,omitempty"`
	ViewURL          string                 `json:"view_url,omitempty"`
	DownloadURL      string                 `json:"download_url,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

func toDocumentResponse(doc db.Document, objectURL, viewURL, downloadURL string) DocumentResponse {
	var metadata map[string]interface{}
	if len(doc.Metadata) > 0 {
		_ = json.Unmarshal(doc.Metadata, &metadata)
	}
	return DocumentResponse{
		ID:               doc.ID,
		Metadata:         metadata,
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
	db *sql.DB,
	remoteDB *sql.DB,
) *DocumentHandler {
	return &DocumentHandler{
		documentService:        documentService,
		auditService:           auditService,
		storageLocationService: storageLocationService,
		storage:                storage,
		storageFactory:         storageFactory,
		db:                     db,
		remoteDB:               remoteDB,
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

func (h *DocumentHandler) GetDocumentStats(c *gin.Context) {
	ctx := c.Request.Context()

	row := h.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)                                              AS total,
			COALESCE(SUM(size_bytes), 0)                         AS total_size,
			COUNT(*) FILTER (WHERE status = 'PENDING')           AS pending,
			COUNT(*) FILTER (WHERE status = 'PROCESSING')        AS processing,
			COUNT(*) FILTER (WHERE status = 'COMPLETED')         AS completed,
			COUNT(*) FILTER (WHERE status = 'FAILED')            AS failed
		FROM documents
		WHERE is_template = false
	`)

	var stats struct {
		Total      int64 `json:"total"`
		TotalSize  int64 `json:"total_size"`
		Pending    int64 `json:"pending"`
		Processing int64 `json:"processing"`
		Completed  int64 `json:"completed"`
		Failed     int64 `json:"failed"`
	}

	if err := row.Scan(
		&stats.Total, &stats.TotalSize,
		&stats.Pending, &stats.Processing,
		&stats.Completed, &stats.Failed,
	); err != nil {
		response.Fail(c, http.StatusInternalServerError, "STATS_FAILED", "failed to compute stats")
		return
	}

	response.OK(c, http.StatusOK, stats)
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", "1000")
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

// ── Data Preview ──────────────────────────────────────────────────────────────

type previewSheet struct {
	Name     string                   `json:"name"`
	RowCount int                      `json:"row_count"`
	Columns  []string                 `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
}

type dataPreviewResponse struct {
	TemplateCode string         `json:"template_code"`
	ReportDate   string         `json:"report_date,omitempty"`
	Sheets       []previewSheet `json:"sheets"`
}

var previewSheetDefs = map[string][]struct {
	name    string
	table   string
	columns []string
}{
	"JMS_STOCK_REPORT": {
		{
			name:  "JMS Stock Issues",
			table: "import.jms_stock_issues",
			columns: []string{
				"report_date", "order_no", "sell_to_customer_no", "customer_name",
				"district", "reference", "part_no", "part_description",
				"owning_customer_no", "owning_customer_name", "uom", "item_status",
				"sold_qty", "desired_qty", "unit_price", "date_sold",
				"association_no", "expiry_dates", "ofr",
			},
		},
		{
			name:  "JMS Stock On Hand",
			table: "import.jms_stock_on_hand",
			columns: []string{
				"report_date", "location_code", "location_name", "bin_code",
				"receipt_date", "receipt_no", "purchase_order_no", "no",
				"description", "lot_no", "expiration_date", "available_quantity",
				"unit_of_measure_code", "unit_cost", "ownership", "total_cost",
				"transfer_order_no", "from_location", "to_location",
				"posting_description", "your_reference",
			},
		},
	},
	"NMS_STOCK_REPORT": {
		{
			name:  "NMS Stock Issues",
			table: "import.nms_stock_issues",
			columns: []string{
				"report_date", "order_type", "ship_to_facility_code", "ship_to_facility_name",
				"district", "item_code", "item_description", "unit_of_measure",
				"order_number", "order_quantity", "quantity_shipped",
				"lot_number", "lot_expiration_date", "list_price", "selling_price",
				"amount_ugx", "ship_confirm_date",
			},
		},
		{
			name:  "NMS Stock On Hand",
			table: "import.nms_stock_on_hand",
			columns: []string{
				"report_date", "item_code", "item_description", "item_inventory_status",
				"funding_source", "subinventory", "locator", "lot_number",
				"expiry_date", "uom", "on_hand_quantity",
			},
		},
	},
	"GF_PIPELINE": {
		{
			name:  "Track and Trace",
			table: "import.gf_pipeline",
			columns: []string{
				"report_date", "psa_name", "country", "grant_name", "grant_budget_id",
				"req_number", "epo_number", "po_number", "shipment_number", "status",
				"vendor_group_name", "item_name_tgf", "sq_so_item_quantity",
				"inco_term_client", "shipment_mode", "country_ship_to_city",
				"po_item_quantity", "shipment_item_qty_ordered", "confirmed_received_quantity",
				"number_of_pallets", "shipment_gross_weight", "shipment_volume",
				"number_of_containers_total", "estimated_vendor_inco_date",
				"estimated_delivery_date", "delivery_date_actual", "so_item_linenumber",
			},
		},
	},
	"GDF_TB_ORDERS": {
		{
			name:  "TB Orders",
			table: "import.gdf_tb_orders",
			columns: []string{
				"report_date", "year_created", "country", "line", "serial_number",
				"total_cost", "order_status", "shipment_code", "product_code",
				"inn_code", "supplier", "quantity_shipped", "units_shipped", "price",
				"estimated_delivery", "actual_delivery", "inco_term", "shipping_mode",
			},
		},
	},
	"GHSC_PSM": {
		{
			name:  "LAB",
			table: "import.ghsc_psm_lab",
			columns: []string{
				"report_date", "line_number", "commodity_category", "item_description",
				"uom", "stock_on_hand", "amc", "mos", "quantity_on_order",
				"requested_delivery_date", "estimated_delivery_date", "status", "comment",
			},
		},
		{
			name:  "Pharma",
			table: "import.ghsc_psm_pharma",
			columns: []string{
				"report_date", "line_number", "commodity_category", "item_description",
				"uom", "stock_on_hand", "amc", "mos", "quantity_on_order",
				"estimated_delivery_date", "status",
			},
		},
		{
			name:  "Malaria",
			table: "import.ghsc_psm_malaria",
			columns: []string{
				"report_date", "line_number", "commodity_category", "item_description",
				"uom", "stock_on_hand", "amc", "mos", "quantity_on_order",
				"estimated_delivery_date", "status",
			},
		},
	},
	"UNFPA": {
		{
			name:  "UNFPA Pipeline",
			table: "import.unfpa_pipeline",
			columns: []string{
				"report_date", "requisition_no", "dept", "product_id", "quantum_item_number",
				"uom", "moh_units", "dkt_units", "msi_units", "psi_units", "ippf_units",
				"total_units", "total_cost", "unit_price", "vendor", "req_line_no",
				"po_number", "po_due_date", "order_life_cycle", "fund_status", "status",
				"eta", "tranche", "funding_year", "period",
			},
		},
		{
			name:  "UNFPA NMS Pipeline",
			table: "import.unfpa_nms_pipeline",
			columns: []string{
				"report_date", "item", "item_id", "item_name", "mot", "eta",
				"quantity", "value", "supplier", "po_number",
				"in_production_until", "eta_as_per_offer", "status",
			},
		},
	},
}

func (h *DocumentHandler) DataPreview(c *gin.Context) {
	ctx := c.Request.Context()

	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid document id")
		return
	}

	doc, err := h.documentService.GetDocument(ctx, documentID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	var meta map[string]string
	if len(doc.Metadata) > 0 {
		_ = json.Unmarshal(doc.Metadata, &meta)
	}

	templateCode := strings.ToUpper(strings.TrimSpace(meta["template_code"]))
	reportDate := meta["report_date"]

	if templateCode == "" {
		response.OK(c, http.StatusOK, dataPreviewResponse{Sheets: []previewSheet{}})
		return
	}

	sheetDefs, ok := previewSheetDefs[templateCode]
	if !ok {
		response.OK(c, http.StatusOK, dataPreviewResponse{TemplateCode: templateCode, Sheets: []previewSheet{}})
		return
	}

	if h.remoteDB == nil {
		response.Fail(c, http.StatusServiceUnavailable, "NO_REMOTE_DB", "remote database not configured")
		return
	}

	result := dataPreviewResponse{TemplateCode: templateCode, ReportDate: reportDate}

	for _, def := range sheetDefs {
		colList := strings.Join(def.columns, ", ")

		// Row count
		var count int
		_ = h.remoteDB.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE document_id = $1 AND is_valid = TRUE`, def.table),
			documentID,
		).Scan(&count)

		// All rows ordered by row_number
		q := fmt.Sprintf(
			`SELECT %s FROM %s WHERE document_id = $1 AND is_valid = TRUE ORDER BY row_number`,
			colList, def.table,
		)
		dbRows, err := h.remoteDB.QueryContext(ctx, q, documentID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "QUERY_FAILED", err.Error())
			return
		}

		colNames, _ := dbRows.Columns()
		var rows []map[string]interface{}
		for dbRows.Next() {
			vals := make([]interface{}, len(colNames))
			ptrs := make([]interface{}, len(colNames))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := dbRows.Scan(ptrs...); err != nil {
				continue
			}
			row := make(map[string]interface{}, len(colNames))
			for i, name := range colNames {
				v := vals[i]
				if b, ok := v.([]byte); ok {
					v = string(b)
				}
				row[name] = v
			}
			rows = append(rows, row)
		}
		dbRows.Close()

		if rows == nil {
			rows = []map[string]interface{}{}
		}

		result.Sheets = append(result.Sheets, previewSheet{
			Name:     def.name,
			RowCount: count,
			Columns:  colNames,
			Rows:     rows,
		})
	}

	response.OK(c, http.StatusOK, result)
}
