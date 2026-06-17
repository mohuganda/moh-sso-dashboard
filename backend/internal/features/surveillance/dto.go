package surveillance

import (
	"time"

	"github.com/google/uuid"
)

type AlertListParams struct {
	EpiWeekID  uuid.UUID
	DiseaseID  uuid.UUID
	DistrictID uuid.UUID
	RegionID   uuid.UUID
}

type AlertResponse struct {
	ID           uuid.UUID  `json:"id"`
	ExternalID   *string    `json:"external_id,omitempty"`
	DiseaseID    uuid.UUID  `json:"disease_id"`
	DiseaseName  *string    `json:"disease_name,omitempty"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty"`
	DistrictName *string    `json:"district_name,omitempty"`
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	EpiWeekID    *uuid.UUID `json:"epi_week_id,omitempty"`
	EpiYear      *int32     `json:"epi_year,omitempty"`
	EpiWeek      *int32     `json:"epi_week,omitempty"`
	OccurredOn   *time.Time `json:"occurred_on,omitempty"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	Narrative    string     `json:"narrative"`
	SubmittedBy  *string    `json:"submitted_by,omitempty"`
	Status       string     `json:"status"`
	SourceName   *string    `json:"source_name,omitempty"`
	ImportedAt   time.Time  `json:"imported_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type WeeklyStatusDetailedResponse struct {
	ID            uuid.UUID  `json:"id"`
	RegionID      *uuid.UUID `json:"region_id,omitempty"`
	DistrictID    *uuid.UUID `json:"district_id,omitempty"`
	SubCountyID   *uuid.UUID `json:"sub_county_id,omitempty"`
	DiseaseID     *uuid.UUID `json:"disease_id,omitempty"`
	IndicatorID   *uuid.UUID `json:"indicator_id,omitempty"`
	EpiWeekID     uuid.UUID  `json:"epi_week_id"`
	Status        string     `json:"status"`
	SourceName    string     `json:"source_name"`
	ImportedAt    time.Time  `json:"imported_at"`
	CreatedAt     time.Time  `json:"created_at"`
	RegionName    string     `json:"region_name"`
	DistrictName  string     `json:"district_name"`
	SubCountyName string     `json:"sub_county_name"`
	DiseaseName   string     `json:"disease_name"`
	IndicatorName string     `json:"indicator_name"`
}

type FacilityWeeklyMetricResponse struct {
	ID             uuid.UUID `json:"id"`
	SourceRecordID *string   `json:"source_record_id,omitempty"`

	FacilityID   uuid.UUID `json:"facility_id"`
	FacilityName string    `json:"facility_name"`

	SubCountyID   *uuid.UUID `json:"sub_county_id,omitempty"`
	SubCountyName string     `json:"sub_county_name"`

	DistrictID   *uuid.UUID `json:"district_id,omitempty"`
	DistrictName string     `json:"district_name"`

	RegionID   *uuid.UUID `json:"region_id,omitempty"`
	RegionName string     `json:"region_name"`

	DiseaseID     *uuid.UUID `json:"disease_id,omitempty"`
	DiseaseName   string     `json:"disease_name"`
	IndicatorID   *uuid.UUID `json:"indicator_id,omitempty"`
	IndicatorName string     `json:"indicator_name"`

	EpiWeekID uuid.UUID `json:"epi_week_id"`
	EpiYear   int32     `json:"epi_year"`
	EpiWeek   int32     `json:"epi_week"`

	MetricValue string    `json:"metric_value"`
	SourceName  string    `json:"source_name"`
	ImportedAt  time.Time `json:"imported_at"`
	CreatedAt   time.Time `json:"created_at"`
}
