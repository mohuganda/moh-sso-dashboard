package surveillance

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

func toAlertResponse(row db.ListAlertsRow) AlertResponse {
	var externalID *string
	if row.ExternalID.Valid {
		externalID = &row.ExternalID.String
	}

	var diseaseName *string
	if row.DiseaseName.Valid {
		diseaseName = &row.DiseaseName.String
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = &row.DistrictID.UUID
	}

	var districtName *string
	if row.DistrictName.Valid {
		districtName = &row.DistrictName.String
	}

	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = &row.RegionID.UUID
	}

	var epiWeekID *uuid.UUID
	if row.EpiWeekID.Valid {
		epiWeekID = &row.EpiWeekID.UUID
	}

	var occurredOn *time.Time
	if row.OccurredOn.Valid {
		occurredOn = &row.OccurredOn.Time
	}

	var createdOn *time.Time
	if row.CreatedOn.Valid {
		createdOn = &row.CreatedOn.Time
	}

	var submittedBy *string
	if row.SubmittedBy.Valid {
		submittedBy = &row.SubmittedBy.String
	}

	var sourceName *string
	if row.SourceName.Valid {
		sourceName = &row.SourceName.String
	}

	return AlertResponse{
		ID:           row.ID,
		ExternalID:   externalID,
		DiseaseID:    row.DiseaseID,
		DiseaseName:  diseaseName,
		DistrictID:   districtID,
		DistrictName: districtName,
		RegionID:     regionID,
		EpiWeekID:    epiWeekID,
		OccurredOn:   occurredOn,
		CreatedOn:    createdOn,
		Narrative:    row.Narrative,
		SubmittedBy:  submittedBy,
		Status:       string(row.Status),
		SourceName:   sourceName,
		ImportedAt:   row.ImportedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toAlertResponses(items []db.ListAlertsRow) []AlertResponse {
	out := make([]AlertResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toAlertResponse(item))
	}
	return out
}

func uuidPtr(v uuid.UUID) *uuid.UUID {
	id := v
	return &id
}

func sqlNullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return strings.TrimSpace(ns.String)
}

func toWeeklyStatusDetailedResponse(row db.ListWeeklyStatusesDetailedRow) WeeklyStatusDetailedResponse {
	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = uuidPtr(row.RegionID.UUID)
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = uuidPtr(row.DistrictID.UUID)
	}

	var subCountyID *uuid.UUID
	if row.SubCountyID.Valid {
		subCountyID = uuidPtr(row.SubCountyID.UUID)
	}

	var diseaseID *uuid.UUID
	if row.DiseaseID.Valid {
		diseaseID = uuidPtr(row.DiseaseID.UUID)
	}

	var indicatorID *uuid.UUID
	if row.IndicatorID.Valid {
		indicatorID = uuidPtr(row.IndicatorID.UUID)
	}

	return WeeklyStatusDetailedResponse{
		ID:            row.ID,
		RegionID:      regionID,
		DistrictID:    districtID,
		SubCountyID:   subCountyID,
		DiseaseID:     diseaseID,
		IndicatorID:   indicatorID,
		EpiWeekID:     row.EpiWeekID,
		Status:        string(row.Status),
		SourceName:    sqlNullStringValue(row.SourceName),
		ImportedAt:    row.ImportedAt,
		CreatedAt:     row.CreatedAt,
		RegionName:    sqlNullStringValue(row.RegionName),
		DistrictName:  sqlNullStringValue(row.DistrictName),
		SubCountyName: sqlNullStringValue(row.SubCountyName),
		DiseaseName:   sqlNullStringValue(row.DiseaseName),
		IndicatorName: sqlNullStringValue(row.IndicatorName),
	}
}

func toFacilityWeeklyDiseaseMetricResponse(row db.ListFacilityWeeklyDiseaseMetricsByWeekRow) FacilityWeeklyMetricResponse {
	var sourceRecordID *string
	if row.SourceRecordID.Valid {
		sourceRecordID = &row.SourceRecordID.String
	}

	var subCountyID *uuid.UUID
	if row.SubCountyID.Valid {
		subCountyID = uuidPtr(row.SubCountyID.UUID)
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = uuidPtr(row.DistrictID.UUID)
	}

	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = uuidPtr(row.RegionID.UUID)
	}

	var diseaseID *uuid.UUID
	if row.DiseaseID.Valid {
		diseaseID = uuidPtr(row.DiseaseID.UUID)
	}

	return FacilityWeeklyMetricResponse{
		ID:             row.ID,
		SourceRecordID: sourceRecordID,
		FacilityID:     row.FacilityID,
		FacilityName:   row.FacilityName,
		SubCountyID:    subCountyID,
		SubCountyName:  sqlNullStringValue(row.SubCountyName),
		DistrictID:     districtID,
		DistrictName:   sqlNullStringValue(row.DistrictName),
		RegionID:       regionID,
		RegionName:     sqlNullStringValue(row.RegionName),
		DiseaseID:      diseaseID,
		DiseaseName:    row.DiseaseName,
		IndicatorID:    nil,
		IndicatorName:  "",
		EpiWeekID:      row.EpiWeekID,
		EpiYear:        row.EpiYear,
		EpiWeek:        row.EpiWeek,
		MetricValue:    row.MetricValue,
		SourceName:     sqlNullStringValue(row.SourceName),
		ImportedAt:     row.ImportedAt,
		CreatedAt:      row.CreatedAt,
	}
}

func toFacilityWeeklyDiseaseMetricByWeekAndDiseaseResponse(row db.ListFacilityWeeklyDiseaseMetricsByWeekAndDiseaseRow) FacilityWeeklyMetricResponse {
	var sourceRecordID *string
	if row.SourceRecordID.Valid {
		sourceRecordID = &row.SourceRecordID.String
	}

	var subCountyID *uuid.UUID
	if row.SubCountyID.Valid {
		subCountyID = uuidPtr(row.SubCountyID.UUID)
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = uuidPtr(row.DistrictID.UUID)
	}

	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = uuidPtr(row.RegionID.UUID)
	}

	var diseaseID *uuid.UUID
	if row.DiseaseID.Valid {
		diseaseID = uuidPtr(row.DiseaseID.UUID)
	}

	return FacilityWeeklyMetricResponse{
		ID:             row.ID,
		SourceRecordID: sourceRecordID,
		FacilityID:     row.FacilityID,
		FacilityName:   row.FacilityName,
		SubCountyID:    subCountyID,
		SubCountyName:  sqlNullStringValue(row.SubCountyName),
		DistrictID:     districtID,
		DistrictName:   sqlNullStringValue(row.DistrictName),
		RegionID:       regionID,
		RegionName:     sqlNullStringValue(row.RegionName),
		DiseaseID:      diseaseID,
		DiseaseName:    row.DiseaseName,
		IndicatorID:    nil,
		IndicatorName:  "",
		EpiWeekID:      row.EpiWeekID,
		EpiYear:        row.EpiYear,
		EpiWeek:        row.EpiWeek,
		MetricValue:    row.MetricValue,
		SourceName:     sqlNullStringValue(row.SourceName),
		ImportedAt:     row.ImportedAt,
		CreatedAt:      row.CreatedAt,
	}
}

func toFacilityWeeklyDiseaseMetricResponses(items []db.ListFacilityWeeklyDiseaseMetricsByWeekRow) []FacilityWeeklyMetricResponse {
	out := make([]FacilityWeeklyMetricResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toFacilityWeeklyDiseaseMetricResponse(item))
	}
	return out
}

func toFacilityWeeklyDiseaseMetricByWeekAndDiseaseResponses(
	rows []db.ListFacilityWeeklyDiseaseMetricsByWeekAndDiseaseRow,
) []FacilityWeeklyMetricResponse {
	out := make([]FacilityWeeklyMetricResponse, 0, len(rows))

	for _, item := range rows {
		out = append(out, toFacilityWeeklyDiseaseMetricByWeekAndDiseaseResponse(item))
	}
	return out
}
