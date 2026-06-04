package service

import (
	"context"
	"testing"
	"time"

	"github.com/moh-sso-dashboard/internal/model"
	"github.com/xuri/excelize/v2"
)

func TestResolveWorkbookSheetNameMatchesNormalizedTemplateIdentifiers(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()

	defaultSheet := workbook.GetSheetName(0)
	if err := workbook.SetSheetName(defaultSheet, "NMS Stock Issues"); err != nil {
		t.Fatalf("rename default sheet: %v", err)
	}

	sheet := model.TemplateSheetStructure{
		Code:        "nms_stock_issues",
		Name:        "nms_stock_issues",
		DisplayName: "NMS Stock Issues",
	}

	got := resolveWorkbookSheetName(workbook, sheet)
	if got != "NMS Stock Issues" {
		t.Fatalf("resolveWorkbookSheetName() = %q, want %q", got, "NMS Stock Issues")
	}
}

func TestResolveWorkbookSheetNamePrefersExactTemplateIdentifier(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()

	defaultSheet := workbook.GetSheetName(0)
	if err := workbook.SetSheetName(defaultSheet, "NMS Stock Issues"); err != nil {
		t.Fatalf("rename default sheet: %v", err)
	}
	if _, err := workbook.NewSheet("nms_stock_issues"); err != nil {
		t.Fatalf("create exact sheet: %v", err)
	}

	sheet := model.TemplateSheetStructure{
		Code:        "nms_stock_issues",
		DisplayName: "NMS Stock Issues",
	}

	got := resolveWorkbookSheetName(workbook, sheet)
	if got != "nms_stock_issues" {
		t.Fatalf("resolveWorkbookSheetName() = %q, want %q", got, "nms_stock_issues")
	}
}

func TestParseValueAcceptsExcelSerialDate(t *testing.T) {
	got, err := parseValue("46111", "DATE")
	if err != nil {
		t.Fatalf("parseValue() returned error: %v", err)
	}

	if got != "2026-03-30" {
		t.Fatalf("parseValue() = %v, want %q", got, "2026-03-30")
	}
}

func TestParseDateToTimeAcceptsTwoDigitSlashDate(t *testing.T) {
	got, err := parseDateToTime("2/27/26")
	if err != nil {
		t.Fatalf("parseDateToTime() returned error: %v", err)
	}

	want := time.Date(2026, time.February, 27, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseDateToTime() = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestExtractSheetRowsRepeatsMergedCellValuesOnContinuationRows(t *testing.T) {
	workbook := excelize.NewFile()
	defer workbook.Close()

	sheetName := workbook.GetSheetName(0)
	for cell, value := range map[string]any{
		"A1": "Commodity Category",
		"B1": "Item description",
		"C1": "UoM",
		"D1": "Quantity on Order",
		"E1": "Status",
		"A2": "OI",
		"B2": "Cotrim 120mg",
		"C2": "Pack",
		"D2": 3380,
		"E2": "Initial order",
		"D3": 3350,
		"E3": "Follow-up order",
		"D4": 1800,
		"E4": "Final order",
	} {
		if err := workbook.SetCellValue(sheetName, cell, value); err != nil {
			t.Fatalf("set %s: %v", cell, err)
		}
	}

	for _, ref := range [][2]string{{"A2", "A4"}, {"B2", "B4"}, {"C2", "C4"}} {
		if err := workbook.MergeCell(sheetName, ref[0], ref[1]); err != nil {
			t.Fatalf("merge %s:%s: %v", ref[0], ref[1], err)
		}
	}

	sheet := model.TemplateSheetStructure{
		Name:      sheetName,
		Required:  true,
		HeaderRow: 1,
		StartRow:  2,
		Columns: []model.TemplateColumnStructure{
			{ColumnKey: "commodity_category", ColumnName: "Commodity Category", DataType: "STRING"},
			{ColumnKey: "item_description", ColumnName: "Item description", DataType: "STRING"},
			{ColumnKey: "uom", ColumnName: "UoM", DataType: "STRING"},
			{ColumnKey: "quantity_on_order", ColumnName: "Quantity on Order", DataType: "DECIMAL"},
			{ColumnKey: "status", ColumnName: "Status", DataType: "STRING"},
		},
	}

	got, err := (&ExcelProcessor{}).extractSheetRows(context.Background(), workbook, sheet, nil)
	if err != nil {
		t.Fatalf("extractSheetRows() returned error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("extractSheetRows() returned %d rows, want 3", len(got))
	}

	for i, row := range got {
		assertEqual(t, row.Data["commodity_category"], "OI")
		assertEqual(t, row.Data["item_description"], "Cotrim 120mg")
		assertEqual(t, row.Data["uom"], "Pack")
		assertEqual(t, row.RawData["commodity_category"], "OI")

		wantStatus := []string{"Initial order", "Follow-up order", "Final order"}[i]
		assertEqual(t, row.Data["status"], wantStatus)
	}

	assertEqual(t, got[0].Data["quantity_on_order"], float64(3380))
	assertEqual(t, got[1].Data["quantity_on_order"], float64(3350))
	assertEqual(t, got[2].Data["quantity_on_order"], float64(1800))
}

func TestNormalizeGHSCRowsForwardFillsContinuationRows(t *testing.T) {
	rows := []extractedExcelRow{
		{
			RowNumber: 9,
			Data: map[string]any{
				"col_1":                           "1",
				"commodity_category":              "OI",
				"item_description":                "Cotrim 120mg",
				"uo_m":                            "1000",
				"stock_on_hand":                   "0",
				"average_monthly_consumption_amc": "1408",
				"mos_as_of_end_of_jan_2026":       "0",
				"quantity_on_order":               "3380",
				"estimated_delivery_date":         "46143",
			},
		},
		{
			RowNumber: 10,
			Data: map[string]any{
				"quantity_on_order":       "3350",
				"estimated_delivery_date": "46143",
				"status":                  "Order released",
			},
		},
	}

	got := normalizeGHSCRows(rows)
	second := got[1].Data

	assertEqual(t, second["line_number"], "1")
	assertEqual(t, second["commodity_category"], "OI")
	assertEqual(t, second["item_description"], "Cotrim 120mg")
	assertEqual(t, second["uom"], "1000")
	assertEqual(t, second["stock_on_hand"], "0")
	assertEqual(t, second["amc"], "1408")
	assertEqual(t, second["mos"], "0")
	assertEqual(t, second["quantity_on_order"], "3350")
	assertEqual(t, second["status"], "Order released")
}

func TestExtractSheetRowsAnchorWithMergedDataCells(t *testing.T) {
	// Simulates a GHSC-style sheet that has a small summary table at the top AND
	// merged data cells in the real table below. Both features must work together:
	// the anchor locates the real header row, and merged-cell expansion fills in
	// continuation rows where Category/Item/UoM are merged vertically.
	workbook := excelize.NewFile()
	defer workbook.Close()

	sheetName := workbook.GetSheetName(0)
	for cell, value := range map[string]any{
		// Small summary table (rows 1-3) — should be ignored
		"A1": "Summary",
		"A2": "Total",
		"B2": 999,
		// Real data table header (row 5)
		"A5": "Commodity Category",
		"B5": "Qty on Order",
		"C5": "Status",
		// Row 6: first data row — Category is merged A6:A8
		"A6": "OI",
		"B6": 3380,
		"C6": "Initial order",
		// Row 7: continuation row (A7 merged, only Qty/Status differ)
		"B7": 3350,
		"C7": "Follow-up",
		// Row 8: continuation row
		"B8": 1800,
		"C8": "Final",
	} {
		if err := workbook.SetCellValue(sheetName, cell, value); err != nil {
			t.Fatalf("set %s: %v", cell, err)
		}
	}
	// Merge Category column vertically across the three data rows.
	if err := workbook.MergeCell(sheetName, "A6", "A8"); err != nil {
		t.Fatalf("merge A6:A8: %v", err)
	}

	sheet := model.TemplateSheetStructure{
		Name:      sheetName,
		Required:  true,
		HeaderRow: 1,
		StartRow:  2,
		Configuration: map[string]any{
			"header_anchor": "Commodity Category",
		},
		Columns: []model.TemplateColumnStructure{
			{ColumnKey: "commodity_category", ColumnName: "Commodity Category", DataType: "STRING"},
			{ColumnKey: "quantity_on_order", ColumnName: "Qty on Order", DataType: "DECIMAL"},
			{ColumnKey: "status", ColumnName: "Status", DataType: "STRING"},
		},
	}

	got, err := (&ExcelProcessor{}).extractSheetRows(context.Background(), workbook, sheet, nil)
	if err != nil {
		t.Fatalf("extractSheetRows() returned error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("extractSheetRows() returned %d rows, want 3", len(got))
	}

	// All three rows should carry "OI" from the merged Category cell.
	for i, row := range got {
		assertEqual(t, row.Data["commodity_category"], "OI")
		wantQty := []float64{3380, 3350, 1800}[i]
		wantStatus := []string{"Initial order", "Follow-up", "Final"}[i]
		assertEqual(t, row.Data["quantity_on_order"], wantQty)
		assertEqual(t, row.Data["status"], wantStatus)
	}
}

func TestExtractSheetRowsAnchorBasedHeaderDetection(t *testing.T) {
	// Simulates a sheet with a small summary table at the top followed by the real
	// data table. The header_anchor config should locate the real table's header.
	workbook := excelize.NewFile()
	defer workbook.Close()

	sheetName := workbook.GetSheetName(0)
	for cell, value := range map[string]any{
		// Small summary table (rows 1-3)
		"A1": "Summary",
		"A2": "Total",
		"B2": 100,
		// Real data table header (row 5)
		"A5": "Commodity Category",
		"B5": "Qty on Order",
		"C5": "Status",
		// Real data rows
		"A6": "OI",
		"B6": 3380,
		"C6": "Initial order",
		"A7": "ARV",
		"B7": 5000,
		"C7": "Confirmed",
	} {
		if err := workbook.SetCellValue(sheetName, cell, value); err != nil {
			t.Fatalf("set %s: %v", cell, err)
		}
	}

	sheet := model.TemplateSheetStructure{
		Name:      sheetName,
		Required:  true,
		HeaderRow: 1,
		StartRow:  2,
		Configuration: map[string]any{
			"header_anchor": "Commodity Category",
		},
		Columns: []model.TemplateColumnStructure{
			{ColumnKey: "commodity_category", ColumnName: "Commodity Category", DataType: "STRING"},
			{ColumnKey: "quantity_on_order", ColumnName: "Qty on Order", DataType: "DECIMAL"},
			{ColumnKey: "status", ColumnName: "Status", DataType: "STRING"},
		},
	}

	got, err := (&ExcelProcessor{}).extractSheetRows(context.Background(), workbook, sheet, nil)
	if err != nil {
		t.Fatalf("extractSheetRows() returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("extractSheetRows() returned %d rows, want 2", len(got))
	}

	assertEqual(t, got[0].Data["commodity_category"], "OI")
	assertEqual(t, got[0].Data["quantity_on_order"], float64(3380))
	assertEqual(t, got[0].Data["status"], "Initial order")
	assertEqual(t, got[1].Data["commodity_category"], "ARV")
	assertEqual(t, got[1].Data["quantity_on_order"], float64(5000))
}

func assertEqual(t *testing.T, got any, want any) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
