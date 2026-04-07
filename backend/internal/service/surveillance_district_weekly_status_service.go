package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type districtStatusPayload struct {
	Week     int32    `json:"weeks"`
	Year     int32    `json:"year"`
	District string   `json:"district"`
	Maroon   []string `json:"maroon"`
	Red      []string `json:"red"`
	Yellow   []string `json:"yellow"`
	Green    []string `json:"green"`
}

type rawDistrictStatusPayload struct {
	Week     json.RawMessage `json:"weeks"`
	Year     json.RawMessage `json:"year"`
	District json.RawMessage `json:"district"`
	Maroon   json.RawMessage `json:"maroon"`
	Red      json.RawMessage `json:"red"`
	Yellow   json.RawMessage `json:"yellow"`
	Green    json.RawMessage `json:"green"`
}

type SurveillanceDistrictWeeklyStatusService struct {
	log                      *logger.Logger
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository
	surveillanceImportRepo   interfaces.ImportRepository
}

func NewSurveillanceDistrictWeeklyStatusService(
	log *logger.Logger,
	districtWeeklyStatusRepo interfaces.DistrictWeeklyStatusRepository,
	surveillanceImportRepo interfaces.ImportRepository,
) *SurveillanceDistrictWeeklyStatusService {
	return &SurveillanceDistrictWeeklyStatusService{
		log:                      log,
		districtWeeklyStatusRepo: districtWeeklyStatusRepo,
		surveillanceImportRepo:   surveillanceImportRepo,
	}
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly statuses by epi week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly statuses by epi week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyStatusesByDistrictAndWeek(
	ctx context.Context,
	arg db.ListDistrictStatusesByDistrictAndWeekParams,
) ([]db.ListDistrictStatusesByDistrictAndWeekRow, error) {
	if err := requireUUID("district id", arg.DistrictID); err != nil {
		return nil, err
	}
	if err := requireUUID("epi week id", arg.EpiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(
		ctx,
		"listing district weekly statuses by district and epi week",
		"district_id", arg.DistrictID,
		"epi_week_id", arg.EpiWeekID,
	)

	items, err := s.districtWeeklyStatusRepo.ListByDistrictAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly statuses by district and epi week",
			"district_id", arg.DistrictID,
			"epi_week_id", arg.EpiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyDiseaseStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyDiseaseStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly disease statuses by week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListDiseaseStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly disease statuses by week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ListDistrictWeeklyIndicatorStatusesByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListDistrictWeeklyIndicatorStatusesByWeekRow, error) {
	if err := requireUUID("epi week id", epiWeekID); err != nil {
		return nil, err
	}

	s.log.Debug(ctx, "listing district weekly indicator statuses by week", "epi_week_id", epiWeekID)

	items, err := s.districtWeeklyStatusRepo.ListIndicatorStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(
			ctx,
			"failed to list district weekly indicator statuses by week",
			"epi_week_id", epiWeekID,
			"error", err,
		)
		return nil, err
	}

	return items, nil
}

func (s *SurveillanceDistrictWeeklyStatusService) ProcessDistrictStatus(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	if err := requireUUID("batch id", batchID); err != nil {
		return err
	}

	s.log.Info(ctx, "processing district weekly status batch", "batch_id", batchID)

	batch, err := s.surveillanceImportRepo.GetImportBatchByID(ctx, batchID)
	if err != nil {
		return fmt.Errorf("get import batch: %w", err)
	}

	fallbackYear, err := resolveBatchYear(batch)
	if err != nil {
		return fmt.Errorf("resolve batch year: %w", err)
	}

	return s.districtWeeklyStatusRepo.WithTx(ctx, func(q db.Querier) error {
		rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("list raw rows by batch: %w", err)
		}

		var successRows int32
		var failedRows int32

		for _, rawRow := range rows {
			if err := s.processDistrictStatusRow(ctx, q, rawRow, fallbackYear); err != nil {
				failedRows++

				s.log.Warn(
					ctx,
					"failed to process district status row",
					"batch_id", batchID,
					"raw_row_id", rawRow.ID,
					"row_number", rawRow.RowNumber,
					"error", err,
				)

				if markErr := s.surveillanceImportRepo.MarkRawRowFailed(ctx, rawRow.ID, err.Error()); markErr != nil {
					return fmt.Errorf("mark raw row failed: %w", markErr)
				}
				continue
			}

			successRows++

			if err := s.surveillanceImportRepo.MarkRawRowProcessed(ctx, rawRow.ID); err != nil {
				return fmt.Errorf("mark raw row processed: %w", err)
			}
		}

		if err := s.surveillanceImportRepo.CompleteImportBatch(ctx, batchID, successRows, failedRows); err != nil {
			return fmt.Errorf("complete batch: %w", err)
		}

		if failedRows == 0 {
			if err := s.surveillanceImportRepo.DeleteProcessedRawRows(ctx, batchID); err != nil {
				return fmt.Errorf("delete processed raw rows: %w", err)
			}
		}

		s.log.Info(
			ctx,
			"finished district weekly status batch processing",
			"batch_id", batchID,
			"success_rows", successRows,
			"failed_rows", failedRows,
		)

		return nil
	})
}

func (s *SurveillanceDistrictWeeklyStatusService) processDistrictStatusRow(
	ctx context.Context,
	q db.Querier,
	rawRow db.SurveillanceImportRawRow,
	fallbackYear int32,
) error {
	payload, err := parseDistrictStatusPayload(rawRow.Payload)
	if err != nil {
		return err
	}

	year, err := resolvePayloadOrBatchYear(payload.Year, fallbackYear, true)
	if err != nil {
		return err
	}

	district, err := q.GetDistrictByName(ctx, payload.District)
	if err != nil {
		return fmt.Errorf("district not found: %s", payload.District)
	}

	epiWeek, err := q.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
		EpiYear: year,
		EpiWeek: payload.Week,
	})
	if err != nil {
		return fmt.Errorf("epi week not found: year=%d week=%d", year, payload.Week)
	}

	statusDiseaseMap := map[db.RiskLevel][]string{
		db.RiskLevelMAROON: payload.Maroon,
		db.RiskLevelRED:    payload.Red,
		db.RiskLevelYELLOW: payload.Yellow,
		db.RiskLevelGREEN:  payload.Green,
	}

	for status, diseases := range statusDiseaseMap {
		for _, diseaseName := range normalizeDiseaseNames(diseases) {
			disease, err := q.GetDiseaseByName(ctx, diseaseName)
			if err != nil {
				return fmt.Errorf("disease not found: %s", diseaseName)
			}

			if _, err := s.districtWeeklyStatusRepo.UpsertDiseaseStatus(ctx, db.UpsertDistrictWeeklyDiseaseStatusParams{
				DistrictID: district.ID,
				DiseaseID:  disease.ID,
				EpiWeekID:  epiWeek.ID,
				Status:     status,
			}); err != nil {
				return fmt.Errorf(
					"upsert district weekly disease status for district=%s disease=%s week=%d year=%d: %w",
					payload.District,
					diseaseName,
					payload.Week,
					year,
					err,
				)
			}
		}
	}

	return nil
}

func parseDistrictStatusPayload(rawPayload []byte) (districtStatusPayload, error) {
	var raw rawDistrictStatusPayload
	if err := json.Unmarshal(rawPayload, &raw); err != nil {
		return districtStatusPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	week, err := parseInt32Field(raw.Week, "weeks")
	if err != nil {
		return districtStatusPayload{}, err
	}

	year, err := parseOptionalInt32Field(raw.Year)
	if err != nil {
		return districtStatusPayload{}, fmt.Errorf("year has invalid format: %w", err)
	}

	district, err := parseStringField(raw.District, "district")
	if err != nil {
		return districtStatusPayload{}, err
	}

	maroon, err := parseStringSliceField(raw.Maroon, "maroon")
	if err != nil {
		return districtStatusPayload{}, err
	}

	red, err := parseStringSliceField(raw.Red, "red")
	if err != nil {
		return districtStatusPayload{}, err
	}

	yellow, err := parseStringSliceField(raw.Yellow, "yellow")
	if err != nil {
		return districtStatusPayload{}, err
	}

	green, err := parseStringSliceField(raw.Green, "green")
	if err != nil {
		return districtStatusPayload{}, err
	}

	payload := districtStatusPayload{
		Week:     week,
		Year:     year,
		District: strings.TrimSpace(district),
		Maroon:   normalizeDiseaseNames(maroon),
		Red:      normalizeDiseaseNames(red),
		Yellow:   normalizeDiseaseNames(yellow),
		Green:    normalizeDiseaseNames(green),
	}

	if payload.District == "" {
		return districtStatusPayload{}, errors.New("district is required")
	}

	if payload.Week <= 0 {
		return districtStatusPayload{}, errors.New("weeks is required")
	}

	if len(payload.Maroon) == 0 &&
		len(payload.Red) == 0 &&
		len(payload.Yellow) == 0 &&
		len(payload.Green) == 0 {
		return districtStatusPayload{}, errors.New("at least one disease status bucket is required")
	}

	return payload, nil
}

func parseInt32Field(raw json.RawMessage, fieldName string) (int32, error) {
	raw = bytesTrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}

	var asInt int32
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return asInt, nil
	}

	var asFloat float64
	if err := json.Unmarshal(raw, &asFloat); err == nil {
		return int32(asFloat), nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return 0, fmt.Errorf("%s is required", fieldName)
		}

		n, err := strconv.ParseFloat(asString, 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be a valid number", fieldName)
		}

		return int32(n), nil
	}

	return 0, fmt.Errorf("%s must be a valid number", fieldName)
}

func parseStringField(raw json.RawMessage, field string) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", fmt.Errorf("%s is required", field)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%s must be a string", field)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}

	return value, nil
}

func parseStringSliceField(raw json.RawMessage, field string) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return normalizeDiseaseNames(arr), nil
	}

	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return nil, fmt.Errorf("%s has invalid format", field)
	}

	str = strings.TrimSpace(str)
	if str == "" || str == "[]" {
		return nil, nil
	}

	normalized := normalizeArrayString(str)
	if normalized == "[]" {
		return nil, nil
	}

	if err := json.Unmarshal([]byte(normalized), &arr); err != nil {
		return nil, fmt.Errorf("%s has invalid list format: %w", field, err)
	}

	return normalizeDiseaseNames(arr), nil
}

func normalizeArrayString(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, `'`, `"`)
	return value
}

func normalizeDiseaseNames(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}

	return result
}

func requireUUID(name string, value uuid.UUID) error {
	if value == uuid.Nil {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func parseOptionalInt32Field(raw json.RawMessage) (int32, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}

	var num int32
	if err := json.Unmarshal(raw, &num); err == nil {
		return num, nil
	}

	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		str = strings.TrimSpace(str)
		if str == "" {
			return 0, nil
		}

		value, convErr := strconv.ParseInt(str, 10, 32)
		if convErr != nil {
			return 0, convErr
		}
		return int32(value), nil
	}

	return 0, fmt.Errorf("unsupported value")
}

func resolvePayloadOrBatchYear(payloadYear, batchYear int32, allowCurrentYearFallback bool) (int32, error) {
	if payloadYear > 0 {
		return payloadYear, nil
	}

	if batchYear > 0 {
		return batchYear, nil
	}

	if allowCurrentYearFallback {
		return int32(time.Now().Year()), nil
	}

	return 0, errors.New("year is required")
}

func resolveBatchYear(batch db.SurveillanceImportBatch) (int32, error) {

	candidates := []string{
		strings.TrimSpace(batch.FileName.String),
		strings.TrimSpace(batch.SourceName),
	}

	if batch.Notes.Valid {
		candidates = append(candidates, strings.TrimSpace(batch.Notes.String))
	}

	for _, candidate := range candidates {
		if year := extractYear(candidate); year > 0 {
			return year, nil
		}
	}

	return 0, nil
}

func extractYear(value string) int32 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	for year := 2035; year >= 2000; year-- {
		token := strconv.Itoa(year)
		if strings.Contains(value, token) {
			return int32(year)
		}
	}

	return 0
}
