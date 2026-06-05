package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	documenttemplates "github.com/moh-sso-dashboard/internal/features/document_templates"
	documentRepository "github.com/moh-sso-dashboard/internal/features/documents"
	"github.com/moh-sso-dashboard/internal/model"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/xuri/excelize/v2"
)

type ExcelProcessor struct {
	documentRepository      documentRepository.DocumentRepository
	documentTemplateService documenttemplates.Service
	stockImportRepository   documentRepository.StockImportRepository
	processRepository       processRepository.ProcessRepository
	storage                 storage.Storage
	remoteDB                *sql.DB
}

func NewExcelProcessor(
	documentRepository documentRepository.DocumentRepository,
	stockImportRepository documentRepository.StockImportRepository,
	processRepository processRepository.ProcessRepository,
	documentTemplateService documenttemplates.Service,
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

	for _, sheet := range template.Sheets {
		sheetCode := strings.ToLower(strings.TrimSpace(sheet.Code))

		if !isSupportedStockSheet(templateCode, sheetCode) {
			continue
		}

		rows, err := c.extractSheetRows(ctx, workbook, sheet)
		if err != nil {
			return fmt.Errorf("extract sheet %q: %w", sheet.Name, err)
		}

		if len(rows) == 0 {
			continue
		}

		if err := c.insertSheetRows(ctx, tx, templateCode, sheet, document.ID, rows); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit remote tx: %w", err)
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
	rows []extractedExcelRow,
) error {
	sheetCode := strings.ToLower(strings.TrimSpace(sheet.Code))
	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))

	const batchSize = 1000

	switch templateCode {
	case "NMS_STOCK_REPORT":
		switch sheetCode {
		case "nms_stock_issues":
			items := make([]model.NMSStockIssue, 0, len(rows))

			for _, row := range rows {
				items = append(items, mapNMSStockIssue(documentID, row))
			}

			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.InsertNMSStockIssuesBatch(ctx, tx, batch); err != nil {
					return fmt.Errorf("insert NMS stock issues batch: %w", err)
				}
			}

		case "nms_stock_on_hand":
			items := make([]model.NMSStockOnHand, 0, len(rows))

			for _, row := range rows {
				items = append(items, mapNMSStockOnHand(documentID, row))
			}

			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.InsertNMSStockOnHandBatch(ctx, tx, batch); err != nil {
					return fmt.Errorf("insert NMS stock on hand batch: %w", err)
				}
			}

		default:
			return nil
		}

	case "JMS_STOCK_REPORT":
		switch sheetCode {
		case "jms_stock_issues":
			items := make([]model.JMSStockIssue, 0, len(rows))

			for _, row := range rows {
				items = append(items, mapJMSStockIssue(documentID, row))
			}

			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.InsertJMSStockIssuesBatch(ctx, tx, batch); err != nil {
					return fmt.Errorf("insert JMS stock issues batch: %w", err)
				}
			}

		case "jms_stock_on_hand":
			items := make([]model.JMSStockOnHand, 0, len(rows))

			for _, row := range rows {
				items = append(items, mapJMSStockOnHand(documentID, row))
			}

			for _, batch := range chunkSlice(items, batchSize) {
				if err := c.stockImportRepository.InsertJMSStockOnHandBatch(ctx, tx, batch); err != nil {
					return fmt.Errorf("insert JMS stock on hand batch: %w", err)
				}
			}

		default:
			return nil
		}

	default:
		return fmt.Errorf("unsupported stock template code: %s", templateCode)
	}

	return nil
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

	if len(rows) <= headerIndex {
		if sheet.Required {
			return nil, fmt.Errorf("header row %d not found", sheet.HeaderRow)
		}

		return nil, nil
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
			if header == "" {
				continue
			}

			value := ""

			if colIndex < len(row) {
				value = strings.TrimSpace(row[colIndex])
			}

			rawData[header] = value

			column, ok := columnByHeader[header]
			if !ok {
				if !sheet.AllowExtraColumns {
					continue
				}

				data[header] = value
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

func mapNMSStockIssue(documentID uuid.UUID, row extractedExcelRow) model.NMSStockIssue {
	return model.NMSStockIssue{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
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
	}
}

func mapNMSStockOnHand(documentID uuid.UUID, row extractedExcelRow) model.NMSStockOnHand {
	return model.NMSStockOnHand{
		DocumentID:          documentID,
		RowNumber:           row.RowNumber,
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
	}
}

func mapJMSStockIssue(documentID uuid.UUID, row extractedExcelRow) model.JMSStockIssue {
	return model.JMSStockIssue{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
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
	}
}

func mapJMSStockOnHand(documentID uuid.UUID, row extractedExcelRow) model.JMSStockOnHand {
	return model.JMSStockOnHand{
		DocumentID:         documentID,
		RowNumber:          row.RowNumber,
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

	case "date":
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
	formats := []string{
		"02-Jan-06",
		"02-Jan-2006",
		"01/02/2006",
		"1/2/2006",
		"2006-01-02",
		"02/01/2006",
		"2/1/2006",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("invalid date")
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

	formats := []string{
		"2006-01-02",
		"02-Jan-06",
		"02-Jan-2006",
		"01/02/2006",
		"1/2/2006",
		"02/01/2006",
		"2/1/2006",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date")
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
