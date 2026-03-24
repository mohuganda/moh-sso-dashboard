package postgres

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type csvRow struct {
	RecordID  string
	Facility  string
	Region    string
	District  string
	SubCounty string
	Disease   string
	Value     float64
	Year      int32
	Week      int32
}

type ImportRepository struct {
	db db.Store
}

func NewImportRepository(db db.Store) *ImportRepository {
	return &ImportRepository{
		db: db,
	}
}

func (r *ImportRepository) CreateImportBatch(ctx context.Context, arg db.CreateImportBatchParams) (db.SurveillanceImportBatch, error) {
	return r.db.CreateImportBatch(ctx, arg)
}

func (r *ImportRepository) GetImportBatchByID(ctx context.Context, id uuid.UUID) (db.SurveillanceImportBatch, error) {
	return r.db.GetImportBatchByID(ctx, id)
}

func (r *ImportRepository) ListImportBatches(ctx context.Context) ([]db.SurveillanceImportBatch, error) {
	return r.db.ListImportBatches(ctx)
}

func (r *ImportRepository) UpdateImportBatchStatus(ctx context.Context, arg db.UpdateImportBatchStatusParams) (db.SurveillanceImportBatch, error) {
	return r.db.UpdateImportBatchStatus(ctx, arg)
}

func (r *ImportRepository) CreateImportRawRow(ctx context.Context, arg db.CreateImportRawRowParams) (db.SurveillanceImportRawRow, error) {
	return r.db.CreateImportRawRow(ctx, arg)
}

func (r *ImportRepository) ListImportRawRowsByBatch(ctx context.Context, batchID uuid.UUID) ([]db.SurveillanceImportRawRow, error) {
	return r.db.ListImportRawRowsByBatch(ctx, batchID)
}

func (r *ImportRepository) ImportCSV(ctx context.Context,
	reader io.Reader,
	fileName string,
	sourceName string) error {
	return nil
}

// func (r *ImportRepository) ImportCSV(
// 	ctx context.Context,
// 	reader io.Reader,
// 	fileName string,
// 	sourceName string,
// ) error {
// 	csvReader := csv.NewReader(reader)
// 	csvReader.TrimLeadingSpace = true

// 	rows, err := csvReader.ReadAll()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read csv: %w", err)
// 	}

// 	if len(rows) < 2 {
// 		return nil, fmt.Errorf("csv has no data rows")
// 	}

// 	headerIndex := mapHeaders(rows[0])

// 	required := []string{
// 		"facility",
// 		"region",
// 		"district",
// 		"sub_county",
// 		"disease",
// 		"value",
// 		"year",
// 		"weeks",
// 	}

// 	for _, col := range required {
// 		if _, ok := headerIndex[col]; !ok {
// 			return nil, fmt.Errorf("missing required column: %s", col)
// 		}
// 	}

// 	result := &service.SurveillanceImportResult{
// 		FileName:   fileName,
// 		SourceName: sourceName,
// 		RowsRead:   len(rows) - 1,
// 	}

// 	for i := 1; i < len(rows); i++ {
// 		record := rows[i]

// 		row, err := parseCSVRow(record, headerIndex)
// 		if err != nil {
// 			result.RowsSkipped++
// 			continue
// 		}

// 		if err := r.importMetricRow(ctx, row, sourceName); err != nil {
// 			result.RowsSkipped++
// 			continue
// 		}

// 		result.RowsImported++
// 	}

// 	result.Message = fmt.Sprintf("processed %d rows, imported %d, skipped %d", result.RowsRead, result.RowsImported, result.RowsSkipped)

// 	return result, nil
// }

func mapHeaders(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		key := strings.ToLower(strings.TrimSpace(h))
		m[key] = i
	}
	return m
}

func getValue(record []string, headerIndex map[string]int, key string) string {
	idx, ok := headerIndex[key]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseCSVRow(record []string, headerIndex map[string]int) (*csvRow, error) {
	value, err := strconv.ParseFloat(getValue(record, headerIndex, "value"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %w", err)
	}

	year, err := strconv.Atoi(getValue(record, headerIndex, "year"))
	if err != nil {
		return nil, fmt.Errorf("invalid year: %w", err)
	}

	week, err := strconv.Atoi(getValue(record, headerIndex, "weeks"))
	if err != nil {
		return nil, fmt.Errorf("invalid week: %w", err)
	}

	return &csvRow{
		RecordID:  getValue(record, headerIndex, "record_id"),
		Facility:  getValue(record, headerIndex, "facility"),
		Region:    getValue(record, headerIndex, "region"),
		District:  getValue(record, headerIndex, "district"),
		SubCounty: getValue(record, headerIndex, "sub_county"),
		Disease:   getValue(record, headerIndex, "disease"),
		Value:     value,
		Year:      int32(year),
		Week:      int32(week),
	}, nil
}

// func (r *ImportRepository) importMetricRow(
// 	ctx context.Context,
// 	row *csvRow,
// 	sourceName string,
// ) error {
// 	region, err := r.db.GetRegionByName(ctx, row.Region)
// 	if err != nil {
// 		return fmt.Errorf("region not found: %w", err)
// 	}

// 	district, err := r.db.GetDistrictByName(ctx, row.District)
// 	if err != nil {
// 		return fmt.Errorf("district not found: %w", err)
// 	}

// 	subcounty, err := r.db.GetSubcountyByName(ctx, row.SubCounty)
// 	if err != nil {
// 		return fmt.Errorf("subcounty not found: %w", err)
// 	}

// 	disease, err := r.db.GetDiseaseByName(ctx, row.Disease)
// 	if err != nil {
// 		return fmt.Errorf("disease not found: %w", err)
// 	}

// 	epiWeek, err := r.db.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
// 		EpiYear: row.Year,
// 		EpiWeek: row.Week,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("epi week not found: %w", err)
// 	}

// 	facility, err := r.db.GetFacilityByName(ctx, row.Facility)
// 	if err != nil {
// 		return fmt.Errorf("facility not found: %w", err)
// 	}

// 	var diseaseID uuid.NullUUID
// 	var indicatorID uuid.NullUUID

// 	if diseaseErr == nil {
// 		diseaseID = uuid.NullUUID{UUID: disease.ID, Valid: true}
// 		indicatorID = uuid.NullUUID{Valid: false}
// 	} else {
// 		indicator, indicatorErr := r.db.GetIndicatorByName(ctx, row.Disease)
// 		if indicatorErr != nil {
// 			return fmt.Errorf("subject not found as disease or indicator: %s", row.Disease)
// 		}
// 		diseaseID = uuid.NullUUID{Valid: false}
// 		indicatorID = uuid.NullUUID{UUID: indicator.ID, Valid: true}
// 	}

// 	// err = r.db.UpsertFacilityWeeklyDiseaseMetric(ctx, db.UpsertFacilityWeeklyDiseaseMetricParams{
// 	// 	ID:          uuid.New(),
// 	// 	FacilityID:  facility.ID,
// 	// 	RegionID:    region.ID,
// 	// 	DistrictID:  district.ID,
// 	// 	SubcountyID: subcounty.ID,
// 	// 	DiseaseID:   uuid.NullUUID{UUID: disease.ID, Valid: true},
// 	// 	IndicatorID: uuid.NullUUID{Valid: false},
// 	// 	EpiWeekID:   epiWeek.ID,
// 	// 	Value:       row.Value,
// 	// 	SourceName:  toNullText(sourceName),
// 	// })
// 	// if err != nil {
// 	// 	return fmt.Errorf("failed to upsert facility weekly metric: %w", err)
// 	// }

// 	return nil
// }

// func toNullText(value string) string {
// 	return strings.TrimSpace(value)
// }
