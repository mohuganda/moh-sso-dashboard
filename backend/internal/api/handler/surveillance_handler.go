package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
)

type SurveillanceHandler struct {
	epiWeekService               *service.SurveillanceEpiWeekService
	diseaseService               *service.SurveillanceDiseaseService
	locationService              *service.SurveillanceLocationService
	facilityWeeklyMetricsService *service.SurveillanceFacilityWeeklyMetricsService
	districtWeeklyStatusService  *service.SurveillanceDistrictWeeklyStatusService
	regionWeeklyStatusService    *service.SurveillanceRegionWeeklyStatusService
	nationalWeeklyStatusService  *service.SurveillanceNationalWeeklyStatusService
}

func NewSurveillanceHandler(
	epiWeekService *service.SurveillanceEpiWeekService,
	diseaseService *service.SurveillanceDiseaseService,
	locationService *service.SurveillanceLocationService,
	facilityWeeklyMetricsService *service.SurveillanceFacilityWeeklyMetricsService,
	districtWeeklyStatusService *service.SurveillanceDistrictWeeklyStatusService,
	regionWeeklyStatusService *service.SurveillanceRegionWeeklyStatusService,
	nationalWeeklyStatusService *service.SurveillanceNationalWeeklyStatusService,
) *SurveillanceHandler {
	return &SurveillanceHandler{
		epiWeekService:               epiWeekService,
		diseaseService:               diseaseService,
		locationService:              locationService,
		facilityWeeklyMetricsService: facilityWeeklyMetricsService,
		districtWeeklyStatusService:  districtWeeklyStatusService,
		regionWeeklyStatusService:    regionWeeklyStatusService,
		nationalWeeklyStatusService:  nationalWeeklyStatusService,
	}
}

// weeks
func (h *SurveillanceHandler) ListEpiWeeksByYearWeek(c *gin.Context) {
	ctx := c.Request.Context()

	yearStr := c.Query("year")
	if yearStr == "" {
		response.Fail(c, http.StatusBadRequest, "year is required", "nil")
		return
	}

	weekStr := c.Query("week")
	if weekStr == "" {
		response.Fail(c, http.StatusBadRequest, "week is required", "nil")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid year", err.Error())
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid week", err.Error())
		return
	}

	params := db.GetEpiWeekByYearWeekParams{
		EpiYear: int32(year),
		EpiWeek: int32(week),
	}

	data, err := h.epiWeekService.GetEpiWeekByYearWeek(ctx, params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get epi week by year and week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListEpiWeeksByYear(c *gin.Context) {
	ctx := c.Request.Context()

	yearStr := c.Query("year")
	if yearStr == "" {
		response.Fail(c, http.StatusBadRequest, "year is required", "nil")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid year", err.Error())
		return
	}

	data, err := h.epiWeekService.ListEpiWeeksByYear(ctx, int32(year))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list epi weeks by year", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// diseases
func (h *SurveillanceHandler) ListDiseases(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.diseaseService.ListDiseases(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list diseases", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// locations

func (h *SurveillanceHandler) ListSubcountiesByDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	districtIDParam := c.Param("districtID")
	if districtIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "districtID is required", "nil")
		return
	}

	districtID, err := uuid.Parse(districtIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid districtID", err.Error())
		return
	}

	data, err := h.locationService.ListSubcountiesByDistrict(ctx, districtID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list subcounties", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) GetSubcountyByID(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", err.Error())
		return
	}

	data, err := h.locationService.GetSubcountyByID(ctx, id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) UpsertDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	var req db.UpsertDistrictParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.locationService.UpsertDistrict(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert district", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) UpsertSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	var req db.UpsertSubCountyParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.locationService.UpsertSubcounty(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) DeleteSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "nil")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", err.Error())
		return
	}

	if err := h.locationService.DeleteSubcounty(ctx, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to delete subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, gin.H{"deleted": true})
}

// facility weekly metrics
func (h *SurveillanceHandler) ListFacilityWeeklyMetricsByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityWeeklyMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly metrics by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListFacilityWeeklyMetricsByFacility(c *gin.Context) {
	ctx := c.Request.Context()

	facilityIDParam := c.Param("facilityID")
	if facilityIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "facilityID is required", "nil")
		return
	}

	facilityID, err := uuid.Parse(facilityIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityWeeklyMetricsByFacility(ctx, facilityID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly metrics by facility", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// district weekly status

func (h *SurveillanceHandler) ListDistrictWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.districtWeeklyStatusService.ListDistrictWeeklyStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list district weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListDistrictWeeklyStatusesByDistrictAndWeek(c *gin.Context) {
	ctx := c.Request.Context()

	districtIDParam := c.Query("district_id")
	if districtIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "district_id is required", "nil")
		return
	}

	districtID, err := uuid.Parse(districtIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid district_id", err.Error())
		return
	}

	epiWeekIDParam := c.Query("epi_week_id")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epi_week_id is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epi_week_id", err.Error())
		return
	}

	arg := db.ListDistrictStatusesByDistrictAndWeekParams{
		DistrictID: districtID,
		EpiWeekID:  epiWeekID,
	}

	data, err := h.districtWeeklyStatusService.ListDistrictWeeklyStatusesByDistrictAndWeek(ctx, arg)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list district weekly statuses by district and week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// region weekly status

func (h *SurveillanceHandler) ListRegionWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.regionWeeklyStatusService.ListRegionWeeklyStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list region weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// national weekly status

func (h *SurveillanceHandler) ListNationalWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.nationalWeeklyStatusService.ListNationalWeeklyStatusesByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list national weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}
