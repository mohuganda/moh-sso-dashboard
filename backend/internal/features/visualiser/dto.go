package visualiser

type DataElementsRequest struct {
	DataSetID *string `json:"data_set_id" form:"data_set_id"`
}

type DataValuesRequest struct {
	OU           []string `json:"ou"`
	PE           []string `json:"pe"`
	DX           []string `json:"dx"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	OrgunitLevel any      `json:"orgunitLevel"`
}

type DataElementsByThemeRequest struct {
	ThemeID string `json:"theme_id" binding:"required"`
}

type DatasetResponse struct {
	DatasetKey  int64  `json:"dataset_key"`
	DatasetID   string `json:"dataset_id"`
	DisplayName string `json:"display_name"`
	IsCurrent   bool   `json:"is_current"`
	CreateDate  string `json:"create_date"`
}

type DataElementResponse struct {
	DimDataElementMapKey int64  `json:"dim_data_element_map_key"`
	DataElementID        string `json:"data_element_id"`
	DataSetID            string `json:"data_set_id"`
	DataElementShortName string `json:"data_element_short_name"`
	DataElementLongName  string `json:"data_element_long_name"`
	RowVersion           int64  `json:"row_version"`
	IsCurrent            bool   `json:"is_current"`
}

type DataValueRowResponse struct {
	OrgUnitID     string `json:"org_unit_id"`
	DataElementID string `json:"data_element_id"`
	Period        string `json:"period"`
	CategoryCombo string `json:"category_combo"`
	Value         int64  `json:"value"`
	Facility      string `json:"facility"`
	Level         string `json:"level"`
	Region        string `json:"region"`
	District      string `json:"district"`
	SubCounty     string `json:"sub_county"`
	Dataelement   string `json:"dataelement"`
}

type DataValuesResponse struct {
	Rows []DataValueRowResponse `json:"rows"`
}

type ThemeResponse struct {
	ThemeID     string `json:"theme_id"`
	ThemeName   string `json:"theme_name"`
	DatasetName string `json:"dataset_name,omitempty"`
}

type DataElementByThemeResponse struct {
	ThemeCategoryID      int64  `json:"theme_category_id"`
	ThemeID              string `json:"theme_id"`
	ThemeName            string `json:"theme_name"`
	DataElementKey       int64  `json:"data_element_key"`
	DataElementID        string `json:"data_element_id"`
	DataElementShortName string `json:"data_element_short_name"`
}

type HIVSummaryResponse struct {
	Year                      int64   `json:"year"`
	Quarter                   int64   `json:"quarter"`
	TotalTested               float64 `json:"total_tested"`
	TotalHIVPositive          float64 `json:"total_hiv_positive"`
	TotalLinkedCare           float64 `json:"total_linked_care"`
	TotalEnrolledCare         float64 `json:"total_enrolled_care"`
	TotalARTStarts            float64 `json:"total_art_starts"`
	TotalARTWithCD4           float64 `json:"total_art_with_cd4"`
	TotalTXCurr               float64 `json:"total_tx_curr"`
	TotalTX1stLine            float64 `json:"total_tx_1st_line"`
	TotalTX2ndLine            float64 `json:"total_tx_2nd_line"`
	TotalTX3rdPlus            float64 `json:"total_tx_3rd_plus"`
	TotalTBScreened           float64 `json:"total_tb_screened"`
	TotalMalnutritionAssessed float64 `json:"total_malnutrition_assessed"`
}

type HIVTestedResponse struct {
	Year              int64   `json:"year"`
	Quarter           int64   `json:"quarter"`
	TestedHIV         float64 `json:"tested_hiv"`
	TestedHIVPositive float64 `json:"tested_hiv_positive"`
	TotalLinkedCare   float64 `json:"total_linked_care"`
}

type HIVRegimenResponse struct {
	Year            int64   `json:"year"`
	Quarter         int64   `json:"quarter"`
	Actives         float64 `json:"actives"`
	TestedViralLoad float64 `json:"tested_viral_load"`
	TotalSuppressed float64 `json:"total_suppressed"`
}
