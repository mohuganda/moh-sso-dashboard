package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type facilityWeeklyMetricsPayload struct {
	RecordID  string  `json:"record_id"`
	Facility  string  `json:"facility"`
	Region    string  `json:"region"`
	District  string  `json:"district"`
	SubCounty string  `json:"sub_county"`
	Metric    string  `json:"disease"`
	Value     float64 `json:"value"`
	Year      int32   `json:"year"`
	Week      int32   `json:"weeks"`
	EpiWeek   int32   `json:"epi_week"`
}

type rawFacilityWeeklyMetricsPayload struct {
	RecordID  string          `json:"record_id"`
	Facility  string          `json:"facility"`
	Region    string          `json:"region"`
	District  string          `json:"district"`
	SubCounty string          `json:"sub_county"`
	Metric    string          `json:"disease"`
	Value     json.RawMessage `json:"value"`
	Year      json.RawMessage `json:"year"`
	Week      json.RawMessage `json:"weeks"`
	EpiWeek   json.RawMessage `json:"epi_week"`
}

type SurveillanceFacilityWeeklyMetricsService struct {
	log                       *logger.Logger
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository
	surveillanceImportRepo    interfaces.ImportRepository
}

func NewSurveillanceFacilityWeeklyMetricsService(
	log *logger.Logger,
	facilityWeeklyMetricsRepo interfaces.FacilityWeeklyMetricsRepository,
	surveillanceImportRepo interfaces.ImportRepository,
) *SurveillanceFacilityWeeklyMetricsService {
	return &SurveillanceFacilityWeeklyMetricsService{
		log:                       log,
		facilityWeeklyMetricsRepo: facilityWeeklyMetricsRepo,
		surveillanceImportRepo:    surveillanceImportRepo,
	}
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly metrics by epi week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListDiseaseMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly metrics by epi week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyMetricsByFacility(
	ctx context.Context,
	facilityID uuid.UUID,
) ([]db.ListFacilityMetricsByFacilityRow, error) {
	s.log.Debug(ctx, "listing facility weekly metrics by facility", "facility_id", facilityID)

	if facilityID == uuid.Nil {
		return []db.ListFacilityMetricsByFacilityRow{}, errors.New("facility id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListByFacility(ctx, facilityID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly metrics by facility", "facility_id", facilityID, "error", err)
		return []db.ListFacilityMetricsByFacilityRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) UpsertFacilityWeeklyIndicatorMetric(
	ctx context.Context,
	arg db.UpsertFacilityWeeklyIndicatorMetricParams,
) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly indicator metric")

	item, err := s.facilityWeeklyMetricsRepo.UpsertIndicatorMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly indicator metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) GetFacilityMetricBySourceRecordID(
	ctx context.Context,
	sourceRecordID string,
) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "getting facility metric by source record id", "source_record_id", sourceRecordID)

	item, err := s.facilityWeeklyMetricsRepo.GetBySourceRecordID(ctx, sourceRecordID)
	if err != nil {
		s.log.Error(ctx, "failed to get facility metric by source record id", "source_record_id", sourceRecordID, "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyDiseaseMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyDiseaseMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly disease metrics by week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListDiseaseMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly disease metrics by week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListFacilityWeeklyDiseaseMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityWeeklyIndicatorMetricsByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.ListFacilityWeeklyIndicatorMetricsByWeekRow, error) {
	s.log.Debug(ctx, "listing facility weekly indicator metrics by week", "epi_week_id", epiWeekID)

	if epiWeekID == uuid.Nil {
		return []db.ListFacilityWeeklyIndicatorMetricsByWeekRow{}, errors.New("epi week id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListIndicatorMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility weekly indicator metrics by week", "epi_week_id", epiWeekID, "error", err)
		return []db.ListFacilityWeeklyIndicatorMetricsByWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) UpsertFacilityWeeklyDiseaseMetric(
	ctx context.Context,
	arg db.UpsertFacilityWeeklyDiseaseMetricParams,
) (db.FacilityWeeklyMetric, error) {
	s.log.Debug(ctx, "upserting facility weekly disease metric")

	item, err := s.facilityWeeklyMetricsRepo.UpsertDiseaseMetric(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert facility weekly disease metric", "error", err)
		return db.FacilityWeeklyMetric{}, err
	}

	return item, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ListFacilityMetricsByFacility(
	ctx context.Context,
	facilityID uuid.UUID,
) ([]db.ListFacilityMetricsByFacilityRow, error) {
	s.log.Debug(ctx, "listing facility metrics by facility", "facility_id", facilityID)

	if facilityID == uuid.Nil {
		return []db.ListFacilityMetricsByFacilityRow{}, errors.New("facility id is required")
	}

	items, err := s.facilityWeeklyMetricsRepo.ListByFacility(ctx, facilityID)
	if err != nil {
		s.log.Error(ctx, "failed to list facility metrics by facility", "facility_id", facilityID, "error", err)
		return []db.ListFacilityMetricsByFacilityRow{}, err
	}

	return items, nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) ProcessFacilityMetrics(
	ctx context.Context,
	batchID uuid.UUID,
) error {
	if batchID == uuid.Nil {
		return fmt.Errorf("batch id is required")
	}

	rows, err := s.surveillanceImportRepo.ListImportRawRowsByBatch(ctx, batchID)
	if err != nil {
		return fmt.Errorf("list import raw rows by batch: %w", err)
	}

	var successRows int32
	var failedRows int32

	for _, raw := range rows {
		err := s.facilityWeeklyMetricsRepo.WithTx(ctx, func(q db.Querier) error {
			return s.processFacilityMetricRow(ctx, q, raw)
		})

		if err != nil {
			failedRows++

			if markErr := s.surveillanceImportRepo.MarkRawRowFailed(ctx, raw.ID, err.Error()); markErr != nil {
				return fmt.Errorf("mark raw row failed: %w", markErr)
			}

			continue
		}

		successRows++

		if err := s.surveillanceImportRepo.MarkRawRowProcessed(ctx, raw.ID); err != nil {
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

	return nil
}

func (s *SurveillanceFacilityWeeklyMetricsService) processFacilityMetricRow(
	ctx context.Context,
	q db.Querier,
	raw db.SurveillanceImportRawRow,
) error {
	payload, err := parseFacilityWeeklyMetricsPayload(raw.Payload)
	if err != nil {
		return err
	}

	region, err := getOrCreateRegion(ctx, q, payload.Region)
	if err != nil {
		return fmt.Errorf("resolve region %q: %w", payload.Region, err)
	}

	district, err := getOrCreateDistrict(ctx, q, payload.District, region.ID)
	if err != nil {
		return fmt.Errorf("resolve district %q: %w", payload.District, err)
	}

	subCounty, err := getOrCreateSubCounty(ctx, q, payload.SubCounty, district.ID)
	if err != nil {
		return fmt.Errorf("resolve sub county %q: %w", payload.SubCounty, err)
	}

	facility, err := getOrCreateFacility(ctx, q, payload.Facility, district.ID, subCounty.ID)
	if err != nil {
		return fmt.Errorf("resolve facility %q: %w", payload.Facility, err)
	}

	epiWeek, err := getOrCreateEpiWeek(ctx, q, payload.Year, payload.Week)
	if err != nil {
		return fmt.Errorf("resolve epi week year=%d week=%d: %w", payload.Year, payload.Week, err)
	}

	metricName := strings.TrimSpace(payload.Metric)
	if metricName == "" {
		return fmt.Errorf("metric is required")
	}

	disease, diseaseErr := q.GetDiseaseByName(ctx, metricName)
	if diseaseErr == nil {
		_, err = q.UpsertFacilityWeeklyDiseaseMetric(ctx, db.UpsertFacilityWeeklyDiseaseMetricParams{
			FacilityID:  facility.ID,
			DiseaseID:   disease.ID,
			EpiWeekID:   epiWeek.ID,
			MetricValue: strconv.FormatFloat(payload.Value, 'f', -1, 64),
		})
		if err != nil {
			return fmt.Errorf("upsert facility weekly disease metric: %w", err)
		}

		return nil
	}

	if diseaseErr != nil && !errors.Is(diseaseErr, sql.ErrNoRows) {
		return fmt.Errorf("find disease %q: %w", metricName, diseaseErr)
	}

	indicator, indicatorErr := q.GetIndicatorByName(ctx, metricName)
	if indicatorErr == nil {
		_, err = q.UpsertFacilityWeeklyIndicatorMetric(ctx, db.UpsertFacilityWeeklyIndicatorMetricParams{
			FacilityID:  facility.ID,
			IndicatorID: indicator.ID,
			EpiWeekID:   epiWeek.ID,
			MetricValue: strconv.FormatFloat(payload.Value, 'f', -1, 64),
		})
		if err != nil {
			return fmt.Errorf("upsert facility weekly indicator metric: %w", err)
		}

		return nil
	}

	if indicatorErr != nil && !errors.Is(indicatorErr, sql.ErrNoRows) {
		return fmt.Errorf("find indicator %q: %w", metricName, indicatorErr)
	}

	return fmt.Errorf("metric %q was not found as a disease or indicator", metricName)
}

func parseFacilityWeeklyMetricsPayload(rawPayload []byte) (facilityWeeklyMetricsPayload, error) {
	var raw rawFacilityWeeklyMetricsPayload
	if err := json.Unmarshal(rawPayload, &raw); err != nil {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("invalid payload json: %w", err)
	}

	year, err := parseInt32Field(raw.Year, "year")
	if err != nil {
		return facilityWeeklyMetricsPayload{}, err
	}

	week, err := parseInt32Field(raw.Week, "weeks")
	if err != nil && len(bytesTrimSpace(raw.EpiWeek)) > 0 {
		week, err = parseInt32Field(raw.EpiWeek, "epi_week")
	}
	if err != nil {
		return facilityWeeklyMetricsPayload{}, err
	}

	value, err := parseFloat64Field(raw.Value, "value")
	if err != nil {
		return facilityWeeklyMetricsPayload{}, err
	}

	payload := facilityWeeklyMetricsPayload{
		RecordID:  strings.TrimSpace(raw.RecordID),
		Facility:  strings.TrimSpace(raw.Facility),
		Region:    strings.TrimSpace(raw.Region),
		District:  strings.TrimSpace(raw.District),
		SubCounty: strings.TrimSpace(raw.SubCounty),
		Metric:    strings.TrimSpace(raw.Metric),
		Value:     value,
		Year:      year,
		Week:      week,
		EpiWeek:   week,
	}

	if payload.Facility == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("facility is required")
	}

	if payload.Region == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("region is required")
	}

	if payload.District == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("district is required")
	}

	if payload.SubCounty == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("sub_county is required")
	}

	if payload.Metric == "" {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("metric is required")
	}

	if payload.Year <= 0 {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("year is required")
	}

	if payload.Week <= 0 {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("weeks is required")
	}

	if payload.Value < 0 {
		return facilityWeeklyMetricsPayload{}, fmt.Errorf("value must be greater than or equal to 0")
	}

	return payload, nil
}

func parseFloat64Field(raw json.RawMessage, fieldName string) (float64, error) {
	raw = bytesTrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}

	var asFloat float64
	if err := json.Unmarshal(raw, &asFloat); err == nil {
		return asFloat, nil
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

		return n, nil
	}

	return 0, fmt.Errorf("%s must be a valid number", fieldName)
}

func bytesTrimSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func getOrCreateRegion(
	ctx context.Context,
	q db.Querier,
	name string,
) (db.Region, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Region{}, fmt.Errorf("region name is required")
	}

	region, err := q.GetRegionByName(ctx, name)
	if err == nil {
		return region, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return db.Region{}, err
	}

	return q.CreateRegion(ctx, db.CreateRegionParams{
		Name: name,
	})
}

func getOrCreateDistrict(
	ctx context.Context,
	q db.Querier,
	name string,
	regionID uuid.UUID,
) (db.District, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.District{}, fmt.Errorf("district name is required")
	}

	district, err := q.GetDistrictByName(ctx, name)
	if err == nil {
		return district, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return db.District{}, err
	}

	return q.CreateDistrict(ctx, db.CreateDistrictParams{
		Name: name,
		RegionID: uuid.NullUUID{
			UUID:  regionID,
			Valid: regionID != uuid.Nil,
		},
	})
}

func getOrCreateSubCounty(
	ctx context.Context,
	q db.Querier,
	name string,
	districtID uuid.UUID,
) (db.SubCounty, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.SubCounty{}, fmt.Errorf("sub county name is required")
	}

	subCounty, err := q.GetSubcountyByName(ctx, name)
	if err == nil {
		return subCounty, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return db.SubCounty{}, err
	}

	return q.CreateSubCounty(ctx, db.CreateSubCountyParams{
		Name:       name,
		DistrictID: districtID,
	})
}

func getOrCreateFacility(
	ctx context.Context,
	q db.Querier,
	name string,
	districtID uuid.UUID,
	subCountyID uuid.UUID,
) (db.Facility, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return db.Facility{}, fmt.Errorf("facility name is required")
	}

	facility, err := q.GetFacilityByName(ctx, name)
	if err == nil {
		return facility, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return db.Facility{}, err
	}

	return q.CreateFacility(ctx, db.CreateFacilityParams{
		Name: name,
		DistrictID: uuid.NullUUID{
			UUID:  districtID,
			Valid: districtID != uuid.Nil,
		},
		SubCountyID: uuid.NullUUID{
			UUID:  subCountyID,
			Valid: subCountyID != uuid.Nil,
		},
	})
}

func getOrCreateEpiWeek(
	ctx context.Context,
	q db.Querier,
	year int32,
	week int32,
) (db.EpiWeek, error) {
	if year <= 0 {
		return db.EpiWeek{}, fmt.Errorf("epi year is required")
	}

	if week <= 0 {
		return db.EpiWeek{}, fmt.Errorf("epi week is required")
	}

	item, err := q.GetEpiWeekByYearWeek(ctx, db.GetEpiWeekByYearWeekParams{
		EpiYear: year,
		EpiWeek: week,
	})
	if err == nil {
		return item, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return db.EpiWeek{}, err
	}

	return q.CreateEpiWeek(ctx, db.CreateEpiWeekParams{
		EpiYear: year,
		EpiWeek: week,
	})
}
