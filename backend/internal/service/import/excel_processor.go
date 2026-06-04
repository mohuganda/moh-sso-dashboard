package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	documentRepository "github.com/moh-sso-dashboard/internal/repository/document"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/xuri/excelize/v2"
)

type ExcelProcessor struct {
	documentRepository      documentRepository.DocumentRepository
	documentTemplateService service.DocumentTemplateService
	stockImportRepository   documentRepository.StockImportRepository
	processRepository       processRepository.ProcessRepository
	storage                 storage.Storage
	remoteDB                *sql.DB
}

func NewExcelProcessor(
	documentRepository documentRepository.DocumentRepository,
	stockImportRepository documentRepository.StockImportRepository,
	processRepository processRepository.ProcessRepository,
	documentTemplateService service.DocumentTemplateService,
	storage storage.Storage,
	remoteDB *sql.DB,
) *ExcelProcessor {
	return &ExcelProcessor{
		documentRepository:      documentRepository,
		stockImportRepository:   stockImportRepository,
		processRepository:       processRepository,
		documentTemplateService: documentTemplateService,
		storage:                 storage,
		remoteDB:                remoteDB,
	}
}

func (c *ExcelProcessor) Process(ctx context.Context, p db.Process) error {
	document, err := c.documentRepository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	if document.IsTemplate {
		return nil
	}

	templateCode, err := getTemplateCodeFromDocument(document)
	if err != nil {
		return err
	}

	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))

	reportDate, err := getReportDateFromDocument(document)
	if err != nil {
		return err
	}

	replaceDocumentID, err := getReplaceDocumentIDFromDocument(document)
	if err != nil {
		return err
	}

	template, err := c.documentTemplateService.GetTemplateStructure(ctx, templateCode)
	if err != nil {
		return fmt.Errorf("get template structure %q: %w", templateCode, err)
	}

	reader, err := c.storage.Download(ctx, document.ObjectKey)
	if err != nil {
		return fmt.Errorf("download document: %w", err)
	}
	defer reader.Close()

	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read downloaded document: %w", err)
	}

	workbook, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return fmt.Errorf("open excel workbook: %w", err)
	}
	defer workbook.Close()

	tx, err := c.remoteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin remote tx: %w", err)
	}
	defer tx.Rollback()

	// If this is a replace upload, invalidate all records from the previous document first
	if replaceDocumentID != nil {
		if err := c.invalidatePreviousDocument(ctx, tx, templateCode, *replaceDocumentID); err != nil {
			return fmt.Errorf("invalidate previous document: %w", err)
		}
	}

	// sheetHashes tracks hashes per sheet code for post-commit soft-delete
	sheetHashes := make(map[string][]string)

	for _, sheet := range template.Sheets {
		sheetCode := strings.ToLower(strings.TrimSpace(sheet.Code))

		if !isSupportedStockSheet(templateCode, sheetCode) {
			continue
		}

		positionalOverrides := blankHeaderOverrides(templateCode, sheetCode)
		rows, err := c.extractSheetRows(ctx, workbook, sheet, positionalOverrides)
		if err != nil {
			return fmt.Errorf("extract sheet %q: %w", sheet.Name, err)
		}

		if len(rows) == 0 {
			continue
		}

		hashes, err := c.insertSheetRows(ctx, tx, templateCode, sheet, document.ID, reportDate, rows)
		if err != nil {
			return err
		}
		sheetHashes[sheetCode] = hashes
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit remote tx: %w", err)
	}

	// Soft-delete rows from previous uploads that are no longer in this file
	sdTx, err := c.remoteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin soft-delete tx: %w", err)
	}
	defer sdTx.Rollback()

	for sheetCode, hashes := range sheetHashes {
		var sdErr error
		switch templateCode + "/" + sheetCode {
		case "NMS_STOCK_REPORT/nms_stock_issues":
			sdErr = c.stockImportRepository.SoftDeleteNMSStockIssues(ctx, sdTx, hashes)
		case "NMS_STOCK_REPORT/nms_stock_on_hand":
			sdErr = c.stockImportRepository.SoftDeleteNMSStockOnHand(ctx, sdTx, hashes)
		case "JMS_STOCK_REPORT/jms_stock_issues":
			sdErr = c.stockImportRepository.SoftDeleteJMSStockIssues(ctx, sdTx, hashes)
		case "JMS_STOCK_REPORT/jms_stock_on_hand":
			sdErr = c.stockImportRepository.SoftDeleteJMSStockOnHand(ctx, sdTx, hashes)
		case "GF_PIPELINE/track_and_trace":
			sdErr = c.stockImportRepository.SoftDeleteGFPipeline(ctx, sdTx, hashes)
		case "GDF_TB_ORDERS/tb_orders":
			sdErr = c.stockImportRepository.SoftDeleteGDFTBOrders(ctx, sdTx, hashes)
		case "GHSC_PSM/lab":
			sdErr = c.stockImportRepository.SoftDeleteGHSCPSMLab(ctx, sdTx, hashes)
		case "GHSC_PSM/pharma":
			sdErr = c.stockImportRepository.SoftDeleteGHSCPSMPharma(ctx, sdTx, hashes)
		case "GHSC_PSM/malaria":
			sdErr = c.stockImportRepository.SoftDeleteGHSCPSMMalaria(ctx, sdTx, hashes)
		}
		if sdErr != nil {
			return fmt.Errorf("soft-delete %s: %w", sheetCode, sdErr)
		}
	}

	if err := sdTx.Commit(); err != nil {
		return fmt.Errorf("commit soft-delete tx: %w", err)
	}

	return nil
}

func isSupportedStockSheet(templateCode string, sheetCode string) bool {
	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))
	sheetCode = strings.ToLower(strings.TrimSpace(sheetCode))

	switch templateCode {
	case "NMS_STOCK_REPORT":
		return sheetCode == "nms_stock_issues" ||
			sheetCode == "nms_stock_on_hand"

	case "JMS_STOCK_REPORT":
		return sheetCode == "jms_stock_issues" ||
			sheetCode == "jms_stock_on_hand"

	case "GF_PIPELINE":
		return sheetCode == "track_and_trace"

	case "GDF_TB_ORDERS":
		return sheetCode == "tb_orders"

	case "GHSC_PSM":
		return sheetCode == "lab" ||
			sheetCode == "pharma" ||
			sheetCode == "malaria"

	default:
		return false
	}
}

func (c *ExcelProcessor) insertSheetRows(
	ctx context.Context,
	tx *sql.Tx,
	templateCode string,
	sheet model.TemplateSheetStructure,
	documentID uuid.UUID,
	reportDate *time.Time,
	rows []extractedExcelRow,
) ([]string, error) {
	sheetCode := strings.ToLower(strings.TrimSpace(sheet.Code))
	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))

	const batchSize = 1000

	switch templateCode {
	case "NMS_STOCK_REPORT":
		switch sheetCode {
		case "nms_stock_issues":
			items := make([]model.NMSStockIssue, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapNMSStockIssue(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.NMSStockIssue) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertNMSStockIssuesBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert NMS stock issues batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		case "nms_stock_on_hand":
			items := make([]model.NMSStockOnHand, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapNMSStockOnHand(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.NMSStockOnHand) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertNMSStockOnHandBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert NMS stock on hand batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		default:
			return nil, nil
		}

	case "JMS_STOCK_REPORT":
		switch sheetCode {
		case "jms_stock_issues":
			items := make([]model.JMSStockIssue, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapJMSStockIssue(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.JMSStockIssue) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertJMSStockIssuesBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert JMS stock issues batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		case "jms_stock_on_hand":
			items := make([]model.JMSStockOnHand, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapJMSStockOnHand(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.JMSStockOnHand) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertJMSStockOnHandBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert JMS stock on hand batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		default:
			return nil, nil
		}

	case "GF_PIPELINE":
		items := make([]model.GFPipelineRow, 0, len(rows))
		for _, row := range rows {
			items = append(items, mapGFPipelineRow(documentID, reportDate, row))
		}
		items = deduplicateByHash(items, func(it model.GFPipelineRow) string { return it.RowHash })
		for _, batch := range chunkSlice(items, batchSize) {
			if err := c.stockImportRepository.UpsertGFPipelineBatch(ctx, tx, batch); err != nil {
				return nil, fmt.Errorf("upsert GF pipeline batch: %w", err)
			}
		}
		hashes := make([]string, len(items))
		for i, it := range items {
			hashes[i] = it.RowHash
		}
		return hashes, nil

	case "GDF_TB_ORDERS":
		items := make([]model.GDFTBOrder, 0, len(rows))
		for _, row := range rows {
			items = append(items, mapGDFTBOrder(documentID, reportDate, row))
		}
		items = deduplicateByHash(items, func(it model.GDFTBOrder) string { return it.RowHash })
		for _, batch := range chunkSlice(items, batchSize) {
			if err := c.stockImportRepository.UpsertGDFTBOrdersBatch(ctx, tx, batch); err != nil {
				return nil, fmt.Errorf("upsert GDF TB orders batch: %w", err)
			}
		}
		hashes := make([]string, len(items))
		for i, it := range items {
			hashes[i] = it.RowHash
		}
		return hashes, nil

	case "GHSC_PSM":
		rows = normalizeGHSCRows(rows)

		switch sheetCode {
		case "lab":
			items := make([]model.GHSCPSMLab, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapGHSCPSMLab(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.GHSCPSMLab) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertGHSCPSMLabBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert GHSC-PSM lab batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		case "pharma":
			items := make([]model.GHSCPSMCommodity, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapGHSCPSMCommodity(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.GHSCPSMCommodity) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertGHSCPSMPharmaBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert GHSC-PSM pharma batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		case "malaria":
			items := make([]model.GHSCPSMCommodity, 0, len(rows))
			for _, row := range rows {
				items = append(items, mapGHSCPSMCommodity(documentID, reportDate, row))
			}
			items = deduplicateByHash(items, func(it model.GHSCPSMCommodity) string { return it.RowHash })
			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.UpsertGHSCPSMMalariaBatch(ctx, tx, batch); err != nil {
					return nil, fmt.Errorf("upsert GHSC-PSM malaria batch: %w", err)
				}
			}
			hashes := make([]string, len(items))
			for i, it := range items {
				hashes[i] = it.RowHash
			}
			return hashes, nil

		default:
			return nil, nil
		}

	default:
		return nil, fmt.Errorf("unsupported stock template code: %s", templateCode)
	}
}

// deduplicateByHash removes items with duplicate hashes, keeping the first occurrence.
// Required because PostgreSQL's ON CONFLICT DO UPDATE cannot target the same row twice
// within a single INSERT statement.
func deduplicateByHash[T any](items []T, getHash func(T) string) []T {
	seen := make(map[string]struct{}, len(items))
	result := make([]T, 0, len(items))
	for _, item := range items {
		h := getHash(item)
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func chunkSlice[T any](items []T, size int) [][]T {
	if size <= 0 {
		size = 1000
	}

	chunks := make([][]T, 0, (len(items)+size-1)/size)

	for start := 0; start < len(items); start += size {
		end := start + size
		if end > len(items) {
			end = len(items)
		}

		chunks = append(chunks, items[start:end])
	}

	return chunks
}

type extractedExcelRow struct {
	RowNumber int
	Data      map[string]any
	RawData   map[string]any
}

func (c *ExcelProcessor) extractSheetRows(
	ctx context.Context,
	workbook *excelize.File,
	sheet model.TemplateSheetStructure,
	positionalOverrides map[int]string, // 1-based col index → column_key for blank-header columns
) ([]extractedExcelRow, error) {
	_ = ctx

	sheetName := resolveWorkbookSheetName(workbook, sheet)

	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		if sheet.Required {
			return nil, fmt.Errorf("%w (available sheets: %s)", err, strings.Join(workbook.GetSheetList(), ", "))
		}

		return nil, nil
	}

	headerIndex := sheet.HeaderRow - 1
	startIndex := sheet.StartRow - 1

	if headerIndex < 0 {
		headerIndex = 0
	}

	if startIndex <= headerIndex {
		startIndex = headerIndex + 1
	}

	// Anchor-based header detection: when a sheet has multiple tables, scan for the
	// last row containing the anchor value to locate the real data table's header.
	// Configured via sheet.Configuration["header_anchor"] = "<column header value>".
	if anchor, ok := sheet.Configuration["header_anchor"].(string); ok && anchor != "" {
		normalizedAnchor := normalizeHeader(anchor)
		foundAt := -1
		for i, row := range rows {
			for _, cell := range row {
				if normalizeHeader(cell) == normalizedAnchor {
					foundAt = i
					break
				}
			}
		}
		if foundAt >= 0 {
			headerIndex = foundAt
			startIndex = headerIndex + 1
		}
	}

	if len(rows) <= headerIndex {
		if sheet.Required {
			return nil, fmt.Errorf("header row %d not found", sheet.HeaderRow)
		}

		return nil, nil
	}

	// Expand merged cells from the header row onward (not just from startIndex) so that
	// horizontally merged header cells (column group labels) are also filled in.
	rows, err = expandMergedCellValues(workbook, sheetName, rows, headerIndex+1)
	if err != nil {
		return nil, fmt.Errorf("expand merged cells: %w", err)
	}

	headers := normalizeHeaders(rows[headerIndex])
	columnByHeader := buildColumnMap(sheet.Columns)

	result := make([]extractedExcelRow, 0)

	for rowIndex := startIndex; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]

		if isEmptyRow(row) {
			continue
		}

		data := make(map[string]any)
		rawData := make(map[string]any)

		for colIndex, header := range headers {
			value := ""
			if colIndex < len(row) {
				value = strings.TrimSpace(row[colIndex])
			}

			var column model.TemplateColumnStructure
			var ok bool

			if header != "" {
				rawData[header] = value
				column, ok = columnByHeader[header]
			}

			// For blank header cells, use the hardcoded positional override map
			// (e.g. JMS exports suppress certain column headers).
			if !ok && header == "" {
				if colKey, hasOverride := positionalOverrides[colIndex+1]; hasOverride {
					column, ok = columnByHeader[normalizeHeader(colKey)]
					if ok {
						rawData[colKey] = value
					}
				}
			}

			if !ok {
				if header != "" && sheet.AllowExtraColumns {
					data[header] = value
				}
				continue
			}

			parsedValue, err := parseValue(value, column.DataType)
			if err != nil {
				return nil, fmt.Errorf(
					"sheet=%s row=%d column=%s value=%q: %w",
					sheet.Name,
					rowIndex+1,
					column.ColumnName,
					value,
					err,
				)
			}

			data[column.ColumnKey] = parsedValue
		}

		result = append(result, extractedExcelRow{
			RowNumber: rowIndex + 1,
			Data:      data,
			RawData:   rawData,
		})
	}

	return result, nil
}

func expandMergedCellValues(
	workbook *excelize.File,
	sheetName string,
	rows [][]string,
	minRow int,
) ([][]string, error) {
	if minRow < 1 {
		minRow = 1
	}

	rowHasOwnData := make([]bool, len(rows))
	for i, row := range rows {
		rowHasOwnData[i] = !isEmptyRow(row)
	}

	mergedCells, err := workbook.GetMergeCells(sheetName)
	if err != nil {
		return rows, err
	}

	for _, mergedCell := range mergedCells {
		value := strings.TrimSpace(mergedCell.GetCellValue())
		if value == "" {
			continue
		}

		startCol, startRow, err := excelize.CellNameToCoordinates(mergedCell.GetStartAxis())
		if err != nil {
			return rows, err
		}

		endCol, endRow, err := excelize.CellNameToCoordinates(mergedCell.GetEndAxis())
		if err != nil {
			return rows, err
		}

		if startRow > endRow {
			startRow, endRow = endRow, startRow
		}
		if startCol > endCol {
			startCol, endCol = endCol, startCol
		}
		if endRow < minRow {
			continue
		}
		if startRow < minRow {
			startRow = minRow
		}

		for rowNumber := startRow; rowNumber <= endRow; rowNumber++ {
			rowIndex := rowNumber - 1
			if rowIndex < 0 || rowIndex >= len(rows) || !rowHasOwnData[rowIndex] {
				continue
			}

			if len(rows[rowIndex]) < endCol {
				expanded := make([]string, endCol)
				copy(expanded, rows[rowIndex])
				rows[rowIndex] = expanded
			}

			for colNumber := startCol; colNumber <= endCol; colNumber++ {
				colIndex := colNumber - 1
				if strings.TrimSpace(rows[rowIndex][colIndex]) == "" {
					rows[rowIndex][colIndex] = value
				}
			}
		}
	}

	return rows, nil
}

func resolveWorkbookSheetName(
	workbook *excelize.File,
	sheet model.TemplateSheetStructure,
) string {
	candidates := sheetNameCandidates(sheet)

	visibleSheets := make([]string, 0)

	for _, sheetName := range workbook.GetSheetList() {
		visible, err := workbook.GetSheetVisible(sheetName)
		if err != nil {
			continue
		}

		if visible {
			visibleSheets = append(visibleSheets, sheetName)
		}
	}

	for _, candidate := range candidates {
		for _, availableSheet := range visibleSheets {
			if candidate == strings.TrimSpace(availableSheet) {
				return availableSheet
			}
		}
	}

	availableByNormalizedName := make(map[string]string, len(visibleSheets))

	for _, availableSheet := range visibleSheets {
		normalized := normalizeSheetName(availableSheet)
		if normalized == "" {
			continue
		}

		if _, exists := availableByNormalizedName[normalized]; !exists {
			availableByNormalizedName[normalized] = availableSheet
		}
	}

	for _, candidate := range candidates {
		if availableSheet, ok := availableByNormalizedName[normalizeSheetName(candidate)]; ok {
			return availableSheet
		}
	}

	// Fallback 1: workbook sheet name starts with the candidate
	// (handles "(Visible Columns Only)" suffixes, etc.)
	for _, candidate := range candidates {
		normalizedCandidate := normalizeSheetName(candidate)
		if normalizedCandidate == "" {
			continue
		}
		for normalizedSheet, availableSheet := range availableByNormalizedName {
			if strings.HasPrefix(normalizedSheet, normalizedCandidate) {
				return availableSheet
			}
		}
	}

	// Fallback 2: workbook sheet name contains the candidate
	// (handles "TO1 LAB" containing "lab", "MALARIA PIPELINE" containing "malaria", etc.)
	for _, candidate := range candidates {
		normalizedCandidate := normalizeSheetName(candidate)
		if normalizedCandidate == "" {
			continue
		}
		for normalizedSheet, availableSheet := range availableByNormalizedName {
			if strings.Contains(normalizedSheet, normalizedCandidate) {
				return availableSheet
			}
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}

	return ""
}

func sheetNameCandidates(sheet model.TemplateSheetStructure) []string {
	values := []string{
		sheet.Code,
		sheet.Name,
		sheet.DisplayName,
	}

	// Read optional aliases from configuration["aliases"]
	if aliases, ok := sheet.Configuration["aliases"]; ok {
		switch v := aliases.(type) {
		case []any:
			for _, a := range v {
				if s, ok := a.(string); ok {
					values = append(values, s)
				}
			}
		case []string:
			values = append(values, v...)
		}
	}

	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		candidate := strings.TrimSpace(value)
		if candidate == "" {
			continue
		}

		normalized := normalizeSheetName(candidate)
		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		result = append(result, candidate)
	}

	return result
}

func normalizeSheetName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))

	var builder strings.Builder
	builder.Grow(len(value))

	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func hashRawData(rawData map[string]any) string {
	b, _ := json.Marshal(rawData)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func mapNMSStockIssue(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.NMSStockIssue {
	return model.NMSStockIssue{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
		ReportDate:         reportDate,
		OrderType:          getString(row.Data, "order_type"),
		ShipToFacilityCode: getString(row.Data, "ship_to_facility_code"),
		ShipToFacilityName: getString(row.Data, "ship_to_facility_name"),
		BillToFacilityCode: getString(row.Data, "bill_to_facility_code"),
		BillToFacilityName: getString(row.Data, "bill_to_facility_name"),
		District:           getString(row.Data, "district"),
		LOC:                getString(row.Data, "loc"),
		FundingSource:      getString(row.Data, "funding_source"),
		Cycle:              getString(row.Data, "cycle"),
		ItemCode:           getString(row.Data, "item_code"),
		ItemDescription:    getString(row.Data, "item_description"),
		UnitOfMeasure:      getString(row.Data, "unit_of_measure"),
		OrderNumber:        getString(row.Data, "order_number"),
		OrderQuantity:      getFloatPtr(row.Data, "order_quantity"),
		LineID:             getString(row.Data, "line_id"),
		QuantityShipped:    getFloatPtr(row.Data, "quantity_shipped"),
		LotNumber:          getString(row.Data, "lot_number"),
		LotExpirationDate:  getTimePtr(row.Data, "lot_expiration_date"),
		ListPrice:          getFloatPtr(row.Data, "list_price"),
		SellingPrice:       getFloatPtr(row.Data, "selling_price"),
		CurrencyCode:       getString(row.Data, "currency_code"),
		AmountUGX:          getFloatPtr(row.Data, "amount_ugx"),
		ShipToLocation:     getString(row.Data, "ship_to_location"),
		ShipConfirmDate:    getTimePtr(row.Data, "ship_confirm_date"),
		ShipConfirmedBy:    getString(row.Data, "ship_confirmed_by"),
		RawPayload:         row.RawData,
		RowHash:            hashRawData(row.RawData),
	}
}

func mapNMSStockOnHand(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.NMSStockOnHand {
	return model.NMSStockOnHand{
		DocumentID:          documentID,
		RowNumber:           row.RowNumber,
		ReportDate:          reportDate,
		ItemCode:            getString(row.Data, "item_code"),
		ItemDescription:     getString(row.Data, "item_description"),
		ItemInventoryStatus: getString(row.Data, "item_inventory_status"),
		FundingSource:       getString(row.Data, "funding_source"),
		SubInventory:        getString(row.Data, "subinventory"),
		Locator:             getString(row.Data, "locator"),
		LotNumber:           getString(row.Data, "lot_number"),
		ExpiryDate:          getTimePtr(row.Data, "expiry_date"),
		PalletID:            getString(row.Data, "pallet_id"),
		UOM:                 getString(row.Data, "uom"),
		OnHandQuantity:      getFloatPtr(row.Data, "on_hand_quantity"),
		RawPayload:          row.RawData,
		RowHash:             hashRawData(row.RawData),
	}
}

func mapJMSStockIssue(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.JMSStockIssue {
	return model.JMSStockIssue{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
		ReportDate:         reportDate,
		OrderNo:            getString(row.Data, "order_no"),
		SellToCustomerNo:   getString(row.Data, "sell_to_customer_no"),
		CustomerName:       getString(row.Data, "customer_name"),
		District:           getString(row.Data, "district"),
		Reference:          getString(row.Data, "reference"),
		PartNo:             getString(row.Data, "part_no"),
		PartDescription:    getString(row.Data, "part_description"),
		OwningCustomerNo:   getString(row.Data, "owning_customer_no"),
		OwningCustomerName: getString(row.Data, "owning_customer_name"),
		UOM:                getString(row.Data, "uom"),
		ItemStatus:         getString(row.Data, "item_status"),
		SoldQty:            getFloatPtr(row.Data, "sold_qty"),
		DesiredQty:         getFloatPtr(row.Data, "desired_qty"),
		UnitPrice:          getFloatPtr(row.Data, "unit_price"),
		DateSold:           getTimePtr(row.Data, "date_sold"),
		AssociationNo:      getString(row.Data, "association_no"),
		ExpiryDates:        getString(row.Data, "expiry_dates"),
		OFR:                getFloatPtr(row.Data, "ofr"),
		RawPayload:         row.RawData,
		RowHash:            hashRawData(row.RawData),
	}
}

func mapJMSStockOnHand(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.JMSStockOnHand {
	return model.JMSStockOnHand{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
		ReportDate:         reportDate,
		LocationCode:       getString(row.Data, "location_code"),
		LocationName:       getString(row.Data, "location_name"),
		BinCode:            getString(row.Data, "bin_code"),
		ReceiptDate:        getTimePtr(row.Data, "receipt_date"),
		ReceiptNo:          getString(row.Data, "receipt_no"),
		PurchaseOrderNo:    getString(row.Data, "purchase_order_no"),
		No:                 getString(row.Data, "no"),
		Description:        getString(row.Data, "description"),
		LotNo:              getString(row.Data, "lot_no"),
		ExpirationDate:     getTimePtr(row.Data, "expiration_date"),
		AvailableQuantity:  getFloatPtr(row.Data, "available_quantity"),
		UnitOfMeasureCode:  getString(row.Data, "unit_of_measure_code"),
		UnitCost:           getFloatPtr(row.Data, "unit_cost"),
		Ownership:          getString(row.Data, "ownership"),
		TotalCost:          getFloatPtr(row.Data, "total_cost"),
		TransferOrderNo:    getString(row.Data, "transfer_order_no"),
		FromLocation:       getString(row.Data, "from_location"),
		ToLocation:         getString(row.Data, "to_location"),
		PostingDescription: getString(row.Data, "posting_description"),
		YourReference:      getString(row.Data, "your_reference"),
		RawPayload:         row.RawData,
		RowHash:            hashRawData(row.RawData),
	}
}

func mapGFPipelineRow(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.GFPipelineRow {
	return model.GFPipelineRow{
		DocumentID: documentID,
		RowNumber:  row.RowNumber,
		ReportDate: reportDate,
		PSAName:    getString(row.Data, "psa_name"),
		Country:    getString(row.Data, "country"),
		GrantName:  getString(row.Data, "grant_name"),
		// "Grant Budget Identification" → canonical: grant_budget_id, auto-generated: grant_budget_identification
		GrantBudgetID: coalesceString(row.Data, "grant_budget_id", "grant_budget_identification"),
		ReqNumber:     getString(row.Data, "req_number"),
		// "ePO Number" → canonical: epo_number, auto-gen (new normalizeKey): e_po_number
		EPONumber:         coalesceString(row.Data, "epo_number", "e_po_number"),
		PONumber:          getString(row.Data, "po_number"),
		ShipmentNumber:    getString(row.Data, "shipment_number"),
		Status:            getString(row.Data, "status"),
		VendorGroupName:   getString(row.Data, "vendor_group_name"),
		ItemNameTGF:       getString(row.Data, "item_name_tgf"),
		SQSOItemQuantity:  getFloatPtr(row.Data, "sq_so_item_quantity"),
		INCOTermClient:    getString(row.Data, "inco_term_client"),
		ShipmentMode:      getString(row.Data, "shipment_mode"),
		CountryShipToCity: getString(row.Data, "country_ship_to_city"),
		POItemQuantity:    getFloatPtr(row.Data, "po_item_quantity"),
		// "Shipment Item Quantity Ordered" → canonical: shipment_item_qty_ordered, auto-gen: shipment_item_quantity_ordered
		ShipmentItemQtyOrdered:  coalesceFloat(row.Data, "shipment_item_qty_ordered", "shipment_item_quantity_ordered"),
		ConfirmedReceivedQty:    getFloatPtr(row.Data, "confirmed_received_quantity"),
		NumberOfPallets:         getFloatPtr(row.Data, "number_of_pallets"),
		ShipmentGrossWeight:     getFloatPtr(row.Data, "shipment_gross_weight"),
		ShipmentVolume:          getFloatPtr(row.Data, "shipment_volume"),
		NumberOfContainersTotal: getFloatPtr(row.Data, "number_of_containers_total"),
		EstimatedVendorIncoDate: getTimePtr(row.Data, "estimated_vendor_inco_date"),
		// "Estimated Delivery Date(PO / Shipment)" → canonical: estimated_delivery_date, auto-gen: estimated_delivery_datepo_shipment
		EstimatedDeliveryDate: coalesceTime(row.Data, "estimated_delivery_date", "estimated_delivery_datepo_shipment"),
		// "DeliveryDateAsPerPOA / POD - CONF" → canonical: delivery_date_actual, auto-gen (new): delivery_date_as_per_poa_pod_conf, (old): deliverydateasperpoa__pod__conf
		DeliveryDateActual: coalesceTime(row.Data, "delivery_date_actual", "delivery_date_as_per_poa_pod_conf", "deliverydateasperpoa__pod__conf"),
		SOItemLinenumber:   getString(row.Data, "so_item_linenumber"),
		RawPayload:         row.RawData,
		RowHash:            hashRawData(row.RawData),
	}
}

// coalesceString returns the first non-empty string value from data for the given keys in order.
func coalesceString(data map[string]any, keys ...string) string {
	for _, k := range keys {
		if v := getString(data, k); v != "" {
			return v
		}
	}
	return ""
}

// coalesceFloat returns the first non-nil float pointer from data for the given keys in order.
func coalesceFloat(data map[string]any, keys ...string) *float64 {
	for _, k := range keys {
		if v := getFloatPtr(data, k); v != nil {
			return v
		}
	}
	return nil
}

// coalesceTime returns the first non-nil time pointer from data for the given keys in order.
func coalesceTime(data map[string]any, keys ...string) *time.Time {
	for _, k := range keys {
		if v := getTimePtr(data, k); v != nil {
			return v
		}
	}
	return nil
}

func mapGDFTBOrder(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.GDFTBOrder {
	return model.GDFTBOrder{
		DocumentID:        documentID,
		RowNumber:         row.RowNumber,
		ReportDate:        reportDate,
		YearCreated:       getIntPtr(row.Data, "year_created"),
		Country:           getString(row.Data, "country"),
		Line:              getString(row.Data, "line"),
		SerialNumber:      getString(row.Data, "serial_number"),
		TotalCost:         getFloatPtr(row.Data, "total_cost"),
		OrderStatus:       getString(row.Data, "order_status"),
		ShipmentCode:      getString(row.Data, "shipment_code"),
		ProductCode:       getString(row.Data, "product_code"),
		INNCode:           coalesceString(row.Data, "inn_code", "inncode"),
		Supplier:          getString(row.Data, "supplier"),
		QuantityShipped:   getFloatPtr(row.Data, "quantity_shipped"),
		UnitsShipped:      getFloatPtr(row.Data, "units_shipped"),
		Price:             getFloatPtr(row.Data, "price"),
		EstimatedDelivery: getTimePtr(row.Data, "estimated_delivery"),
		ActualDelivery:    getTimePtr(row.Data, "actual_delivery"),
		IncoTerm:          getString(row.Data, "inco_term"),
		ShippingMode:      getString(row.Data, "shipping_mode"),
		RawPayload:        row.RawData,
		RowHash:           hashRawData(row.RawData),
	}
}

func mapGHSCPSMLab(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.GHSCPSMLab {
	return model.GHSCPSMLab{
		DocumentID: documentID,
		RowNumber:  row.RowNumber,
		ReportDate: reportDate,
		// "#" col: normalizeKey("#")="" so stored as data[""]; also check "col_1" for re-uploads
		LineNumber:        coalesceString(row.Data, "line_number", "col_1"),
		CommodityCategory: getString(row.Data, "commodity_category"),
		ItemDescription:   getString(row.Data, "item_description"),
		// "UoM": normalizeKey splits CamelCase → stored as "uo_m"; normalizeHeader → "uom"
		UOM:         coalesceString(row.Data, "uom", "uo_m"),
		StockOnHand: getFloatPtr(row.Data, "stock_on_hand"),
		// "Average Monthly Consumption (AMC)" → stored as "average_monthly_consumption_amc"
		AMC: coalesceFloat(row.Data, "amc", "average_monthly_consumption_amc"),
		// "MOS as of end of January 2026" → stored as "mos_as_of_end_of_january_2026"
		MOS:                   coalesceFloat(row.Data, "mos", "mos_as_of_end_of_january_2026", "mos_as_of_end_of_jan_2026"),
		QuantityOnOrder:       getFloatPtr(row.Data, "quantity_on_order"),
		RequestedDeliveryDate: getTimePtr(row.Data, "requested_delivery_date"),
		EstimatedDeliveryDate: getTimePtr(row.Data, "estimated_delivery_date"),
		Status:                getString(row.Data, "status"),
		Comment:               getString(row.Data, "comment"),
		RawPayload:            row.RawData,
		RowHash:               hashRawData(row.RawData),
	}
}

func mapGHSCPSMCommodity(documentID uuid.UUID, reportDate *time.Time, row extractedExcelRow) model.GHSCPSMCommodity {
	return model.GHSCPSMCommodity{
		DocumentID:        documentID,
		RowNumber:         row.RowNumber,
		ReportDate:        reportDate,
		LineNumber:        coalesceString(row.Data, "line_number", "col_1"),
		CommodityCategory: getString(row.Data, "commodity_category"),
		ItemDescription:   getString(row.Data, "item_description"),
		UOM:               coalesceString(row.Data, "uom", "uo_m"),
		StockOnHand:       getFloatPtr(row.Data, "stock_on_hand"),
		AMC:               coalesceFloat(row.Data, "amc", "average_monthly_consumption_amc"),
		// "MOS as of end of Jan 2026" → "mos_as_of_end_of_jan_2026"
		MOS:                   coalesceFloat(row.Data, "mos", "mos_as_of_end_of_jan_2026", "mos_as_of_end_of_january_2026"),
		QuantityOnOrder:       getFloatPtr(row.Data, "quantity_on_order"),
		EstimatedDeliveryDate: getTimePtr(row.Data, "estimated_delivery_date"),
		Status:                getString(row.Data, "status"),
		RawPayload:            row.RawData,
		RowHash:               hashRawData(row.RawData),
	}
}

func normalizeGHSCRows(rows []extractedExcelRow) []extractedExcelRow {
	carry := make(map[string]any)
	fillKeys := []string{
		"line_number",
		"commodity_category",
		"item_description",
		"uom",
		"stock_on_hand",
		"amc",
		"mos",
	}

	for i := range rows {
		canonicalizeGHSCRowData(rows[i].Data)

		for _, key := range fillKeys {
			if isBlankValue(rows[i].Data[key]) && !isBlankValue(carry[key]) {
				rows[i].Data[key] = carry[key]
			}
		}

		for _, key := range fillKeys {
			if !isBlankValue(rows[i].Data[key]) {
				carry[key] = rows[i].Data[key]
			}
		}
	}

	return rows
}

func canonicalizeGHSCRowData(data map[string]any) {
	if data == nil {
		return
	}

	setCanonicalValue(data, "line_number", "col_1")
	setCanonicalValue(data, "uom", "uo_m")
	setCanonicalValue(data, "amc", "average_monthly_consumption_amc")
	setCanonicalValue(data, "mos", "mos_as_of_end_of_jan_2026", "mos_as_of_end_of_january_2026")
}

func setCanonicalValue(data map[string]any, canonical string, aliases ...string) {
	if !isBlankValue(data[canonical]) {
		return
	}

	for _, alias := range aliases {
		if !isBlankValue(data[alias]) {
			data[canonical] = data[alias]
			return
		}
	}
}

func isBlankValue(value any) bool {
	if value == nil {
		return true
	}

	return strings.TrimSpace(fmt.Sprint(value)) == ""
}

func getIntPtr(data map[string]any, key string) *int {
	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	switch v := value.(type) {
	case int:
		return &v
	case int64:
		n := int(v)
		return &n
	case float64:
		n := int(v)
		return &n
	case string:
		cleaned := strings.TrimSpace(v)
		if cleaned == "" {
			return nil
		}
		n, err := strconv.Atoi(cleaned)
		if err != nil {
			return nil
		}
		return &n
	default:
		return nil
	}
}

func buildColumnMap(columns []model.TemplateColumnStructure) map[string]model.TemplateColumnStructure {
	result := make(map[string]model.TemplateColumnStructure)

	for _, column := range columns {
		result[normalizeHeader(column.ColumnName)] = column
		result[normalizeHeader(column.ColumnKey)] = column
		result[normalizeHeader(column.DisplayName)] = column

		for _, alias := range column.Aliases {
			result[normalizeHeader(alias)] = column
		}
	}

	return result
}

// blankHeaderOverrides returns a 1-based column-index → column_key map for
// known templates where the export suppresses certain column header cells.
// This avoids requiring column_order to be set in the template configuration.
func blankHeaderOverrides(templateCode, sheetCode string) map[int]string {
	switch strings.ToUpper(templateCode) + "/" + strings.ToLower(sheetCode) {
	case "JMS_STOCK_REPORT/jms_stock_issues":
		return map[int]string{
			1: "order_no",
			2: "sell_to_customer_no",
			6: "part_no",
			8: "owning_customer_no",
		}
	case "JMS_STOCK_REPORT/jms_stock_on_hand":
		return map[int]string{
			7:  "no",
			9:  "lot_no",
			14: "ownership",
			16: "transfer_order_no",
		}
	}
	return nil
}

func normalizeHeaders(headers []string) []string {
	result := make([]string, 0, len(headers))

	for _, header := range headers {
		result = append(result, normalizeHeader(header))
	}

	return result
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}

	return true
}

func parseValue(value string, dataType string) (any, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, nil
	}

	switch strings.ToLower(dataType) {
	case "text", "string":
		return value, nil

	case "number", "numeric", "decimal":
		cleaned := strings.ReplaceAll(value, ",", "")
		cleaned = strings.TrimSuffix(cleaned, "%")

		number, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil, err
		}

		return number, nil

	case "integer", "int":
		cleaned := strings.ReplaceAll(value, ",", "")

		number, err := strconv.Atoi(cleaned)
		if err != nil {
			return nil, err
		}

		return number, nil

	case "date", "datetime", "timestamp":
		return parseDate(value)

	case "boolean", "bool":
		switch strings.ToLower(value) {
		case "true", "yes", "1":
			return true, nil
		case "false", "no", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean")
		}

	default:
		return value, nil
	}
}

func parseDate(value string) (string, error) {
	parsed, err := parseDateToTime(value)
	if err != nil {
		return "", err
	}

	return parsed.Format("2006-01-02"), nil
}

func getString(data map[string]any, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}

	return strings.TrimSpace(fmt.Sprint(value))
}

func getFloatPtr(data map[string]any, key string) *float64 {
	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	switch v := value.(type) {
	case float64:
		return &v
	case float32:
		n := float64(v)
		return &n
	case int:
		n := float64(v)
		return &n
	case int64:
		n := float64(v)
		return &n
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		cleaned = strings.TrimSuffix(cleaned, "%")

		if cleaned == "" {
			return nil
		}

		n, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil
		}

		return &n
	default:
		return nil
	}
}

func getTimePtr(data map[string]any, key string) *time.Time {
	value, ok := data[key]
	if !ok || value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		return &v
	case float64:
		return excelSerialDateToTimePtr(v)
	case float32:
		return excelSerialDateToTimePtr(float64(v))
	case int:
		return excelSerialDateToTimePtr(float64(v))
	case int64:
		return excelSerialDateToTimePtr(float64(v))
	case string:
		parsed, err := parseDateToTime(v)
		if err != nil {
			return nil
		}

		return &parsed
	default:
		return nil
	}
}

func parseDateToTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("invalid date")
	}

	if parsed, ok := parseExcelSerialDate(value); ok {
		return parsed, nil
	}

	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"02-Jan-06",
		"02-Jan-2006",
		"2-Jan-06",
		"2-Jan-2006",
		"01/02/2006",
		"1/2/2006",
		"01/02/06",
		"1/2/06",
		// MM-DD-YYYY / MM-DD-YY with dashes (e.g. "02-20-24" from GDF TB Orders)
		"01-02-2006",
		"1-2-2006",
		"01-02-06",
		"1-2-06",
		"02/01/2006",
		"2/1/2006",
		"02/01/06",
		"2/1/06",
		// Month-Year only (e.g. "May-26", "January-2026") — day defaults to 1
		"Jan-06",
		"Jan-2006",
		"January-06",
		"January-2006",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date")
}

func parseExcelSerialDate(value string) (time.Time, bool) {
	cleaned := strings.TrimSpace(strings.ReplaceAll(value, ",", ""))
	serial, err := strconv.ParseFloat(cleaned, 64)
	if err != nil || serial <= 0 {
		return time.Time{}, false
	}

	parsed, err := excelize.ExcelDateToTime(serial, false)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}

func excelSerialDateToTimePtr(serial float64) *time.Time {
	if serial <= 0 {
		return nil
	}

	parsed, err := excelize.ExcelDateToTime(serial, false)
	if err != nil {
		return nil
	}

	return &parsed
}

func getReplaceDocumentIDFromDocument(document db.Document) (*uuid.UUID, error) {
	if len(document.Metadata) == 0 {
		return nil, nil
	}

	var metadata struct {
		ReplaceDocumentID string `json:"replace_document_id"`
	}

	if err := json.Unmarshal(document.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("decode document metadata: %w", err)
	}

	if metadata.ReplaceDocumentID == "" {
		return nil, nil
	}

	id, err := uuid.Parse(metadata.ReplaceDocumentID)
	if err != nil {
		return nil, fmt.Errorf("invalid replace_document_id %q: %w", metadata.ReplaceDocumentID, err)
	}

	return &id, nil
}

func (c *ExcelProcessor) invalidatePreviousDocument(
	ctx context.Context,
	tx *sql.Tx,
	templateCode string,
	replaceDocumentID uuid.UUID,
) error {
	switch strings.ToUpper(templateCode) {
	case "NMS_STOCK_REPORT":
		if err := c.stockImportRepository.InvalidateNMSStockIssuesByDocument(ctx, tx, replaceDocumentID); err != nil {
			return err
		}
		return c.stockImportRepository.InvalidateNMSStockOnHandByDocument(ctx, tx, replaceDocumentID)
	case "JMS_STOCK_REPORT":
		if err := c.stockImportRepository.InvalidateJMSStockIssuesByDocument(ctx, tx, replaceDocumentID); err != nil {
			return err
		}
		return c.stockImportRepository.InvalidateJMSStockOnHandByDocument(ctx, tx, replaceDocumentID)
	case "GF_PIPELINE":
		return c.stockImportRepository.InvalidateGFPipelineByDocument(ctx, tx, replaceDocumentID)
	case "GDF_TB_ORDERS":
		return c.stockImportRepository.InvalidateGDFTBOrdersByDocument(ctx, tx, replaceDocumentID)
	case "GHSC_PSM":
		if err := c.stockImportRepository.InvalidateGHSCPSMLabByDocument(ctx, tx, replaceDocumentID); err != nil {
			return err
		}
		if err := c.stockImportRepository.InvalidateGHSCPSMPharmaByDocument(ctx, tx, replaceDocumentID); err != nil {
			return err
		}
		return c.stockImportRepository.InvalidateGHSCPSMMalariaByDocument(ctx, tx, replaceDocumentID)
	}
	return nil
}

func getReportDateFromDocument(document db.Document) (*time.Time, error) {
	if len(document.Metadata) == 0 {
		return nil, fmt.Errorf("document metadata missing report_date")
	}

	var metadata struct {
		ReportDate string `json:"report_date"`
	}

	if err := json.Unmarshal(document.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("decode document metadata: %w", err)
	}

	if metadata.ReportDate == "" {
		return nil, fmt.Errorf("report_date is required in document metadata")
	}

	parsed, err := time.Parse("2006-01-02", metadata.ReportDate)
	if err != nil {
		return nil, fmt.Errorf("invalid report_date %q: expected YYYY-MM-DD", metadata.ReportDate)
	}

	if parsed.After(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("report_date cannot be a future date")
	}

	return &parsed, nil
}

func getTemplateCodeFromDocument(document db.Document) (string, error) {
	if len(document.Metadata) == 0 {
		return "", fmt.Errorf("document metadata missing template_code")
	}

	var metadata struct {
		TemplateCode string `json:"template_code"`
	}

	if err := json.Unmarshal(document.Metadata, &metadata); err != nil {
		return "", fmt.Errorf("decode document metadata: %w", err)
	}

	templateCode := strings.TrimSpace(metadata.TemplateCode)
	if templateCode == "" {
		return "", fmt.Errorf("template_code is required in document metadata")
	}

	return templateCode, nil
}

func chunkNMSStockIssues(rows []model.NMSStockIssue, size int) [][]model.NMSStockIssue {
	if size <= 0 {
		size = 1000
	}

	chunks := make([][]model.NMSStockIssue, 0, (len(rows)+size-1)/size)

	for start := 0; start < len(rows); start += size {
		end := start + size
		if end > len(rows) {
			end = len(rows)
		}

		chunks = append(chunks, rows[start:end])
	}

	return chunks
}
