package surveillance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
	"github.com/moh-sso-dashboard/internal/utils"
)

type LocationService struct {
	log           *logger.Logger
	regionRepo    RegionRepository
	districtRepo  DistrictRepository
	subCountyRepo SubCountyRepository
	notifications sharedservice.NotificationsService
	cfg           *config.Config
}

func NewLocationService(
	log *logger.Logger,
	regionRepo RegionRepository,
	districtRepo DistrictRepository,
	subCountyRepo SubCountyRepository,
	notifications sharedservice.NotificationsService,
	cfg ...*config.Config,
) *LocationService {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &LocationService{
		log:           log,
		regionRepo:    regionRepo,
		districtRepo:  districtRepo,
		subCountyRepo: subCountyRepo,
		notifications: notifications,
		cfg:           appConfig,
	}
}

// ============================================================
// Regions
// ============================================================

func (s *LocationService) ListRegions(ctx context.Context) ([]db.Region, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return nil, errors.New("region repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing regions")
	}

	items, err := s.regionRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list regions", "error", err)
		}

		return []db.Region{}, fmt.Errorf("list regions: %w", err)
	}

	return items, nil
}

func (s *LocationService) GetRegionByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	if id == uuid.Nil {
		return db.Region{}, errors.New("region id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting region by id", "region_id", id)
	}

	item, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get region by id", "region_id", id, "error", err)
		}

		return db.Region{}, fmt.Errorf("get region by id: %w", err)
	}

	return item, nil
}

func (s *LocationService) GetRegionByName(
	ctx context.Context,
	name string,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return db.Region{}, errors.New("region name is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting region by name", "name", name)
	}

	item, err := s.regionRepo.GetByName(ctx, name)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get region by name", "name", name, "error", err)
		}

		return db.Region{}, fmt.Errorf("get region by name: %w", err)
	}

	return item, nil
}

func (s *LocationService) UpsertRegion(
	ctx context.Context,
	arg db.UpsertRegionParams,
) (db.Region, error) {
	if s == nil {
		return db.Region{}, errors.New("surveillance location service is nil")
	}

	if s.regionRepo == nil {
		return db.Region{}, errors.New("region repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.Region{}, errors.New("region name is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting region", "name", arg.Name)
	}

	item, err := s.regionRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert region", "name", arg.Name, "error", err)
		}

		return db.Region{}, fmt.Errorf("upsert region: %w", err)
	}

	// In-app only. Location upserts can happen during imports.
	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_REGION_UPSERTED",
		Title:      "Region updated",
		Severity:   "info",
		Message:    "Surveillance region created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"region_id": item.ID.String(),
			"name":      item.Name,
		}),
	})

	return item, nil
}

// ============================================================
// Districts
// ============================================================

func (s *LocationService) ListDistricts(
	ctx context.Context,
) ([]db.ListDistrictsRow, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return nil, errors.New("district repository is nil")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing districts")
	}

	items, err := s.districtRepo.List(ctx)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list districts", "error", err)
		}

		return []db.ListDistrictsRow{}, fmt.Errorf("list districts: %w", err)
	}

	return items, nil
}

func (s *LocationService) ListDistrictsByRegion(
	ctx context.Context,
	regionID uuid.UUID,
) ([]db.District, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return nil, errors.New("district repository is nil")
	}

	if regionID == uuid.Nil {
		return nil, errors.New("region id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing districts by region", "region_id", regionID)
	}

	items, err := s.districtRepo.ListByRegion(ctx, regionID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list districts by region", "region_id", regionID, "error", err)
		}

		return []db.District{}, fmt.Errorf("list districts by region: %w", err)
	}

	return items, nil
}

func (s *LocationService) GetDistrictByID(
	ctx context.Context,
	id uuid.UUID,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	if id == uuid.Nil {
		return db.District{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting district by id", "district_id", id)
	}

	item, err := s.districtRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get district by id", "district_id", id, "error", err)
		}

		return db.District{}, fmt.Errorf("get district by id: %w", err)
	}

	return item, nil
}

func (s *LocationService) GetDistrictByName(
	ctx context.Context,
	name string,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return db.District{}, errors.New("district name is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting district by name", "name", name)
	}

	item, err := s.districtRepo.GetByName(ctx, name)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get district by name", "name", name, "error", err)
		}

		return db.District{}, fmt.Errorf("get district by name: %w", err)
	}

	return item, nil
}

func (s *LocationService) UpsertDistrict(
	ctx context.Context,
	arg db.UpsertDistrictParams,
) (db.District, error) {
	if s == nil {
		return db.District{}, errors.New("surveillance location service is nil")
	}

	if s.districtRepo == nil {
		return db.District{}, errors.New("district repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.District{}, errors.New("district name is required")
	}

	if !arg.RegionID.Valid || arg.RegionID.UUID == uuid.Nil {
		return db.District{}, errors.New("region id is required")
	}

	if s.log != nil {
		s.log.Info(
			ctx,
			"upserting district",
			"name", arg.Name,
			"region_id", arg.RegionID.UUID.String(),
		)
	}

	item, err := s.districtRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert district", "name", arg.Name, "error", err)
		}

		return db.District{}, fmt.Errorf("upsert district: %w", err)
	}

	// In-app only. Location upserts can happen during imports.
	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_DISTRICT_UPSERTED",
		Title:      "District updated",
		Severity:   "info",
		Message:    "Surveillance district created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"district_id": item.ID.String(),
			"name":        item.Name,
			"region_id":   nullUUIDString(item.RegionID),
		}),
	})

	return item, nil
}

// ============================================================
// Sub-counties
// ============================================================

func (s *LocationService) ListSubcountiesByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.SubCounty, error) {
	if s == nil {
		return nil, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return nil, errors.New("subcounty repository is nil")
	}

	if districtID == uuid.Nil {
		return []db.SubCounty{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing subcounties by district", "district_id", districtID)
	}

	items, err := s.subCountyRepo.ListByDistrict(ctx, districtID)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list subcounties by district", "district_id", districtID, "error", err)
		}

		return []db.SubCounty{}, fmt.Errorf("list subcounties by district: %w", err)
	}

	return items, nil
}

func (s *LocationService) GetSubcountyByID(
	ctx context.Context,
	id uuid.UUID,
) (db.SubCounty, error) {
	if s == nil {
		return db.SubCounty{}, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return db.SubCounty{}, errors.New("subcounty repository is nil")
	}

	if id == uuid.Nil {
		return db.SubCounty{}, errors.New("subcounty id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting subcounty by id", "id", id)
	}

	item, err := s.subCountyRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get subcounty by id", "id", id, "error", err)
		}

		return db.SubCounty{}, fmt.Errorf("get subcounty by id: %w", err)
	}

	return item, nil
}

func (s *LocationService) UpsertSubcounty(
	ctx context.Context,
	arg db.UpsertSubCountyParams,
) (db.SubCounty, error) {
	if s == nil {
		return db.SubCounty{}, errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return db.SubCounty{}, errors.New("subcounty repository is nil")
	}

	arg.Name = strings.TrimSpace(arg.Name)
	if arg.Name == "" {
		return db.SubCounty{}, errors.New("subcounty name is required")
	}

	if arg.DistrictID == uuid.Nil {
		return db.SubCounty{}, errors.New("district id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting subcounty", "name", arg.Name, "district_id", arg.DistrictID)
	}

	item, err := s.subCountyRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert subcounty", "name", arg.Name, "district_id", arg.DistrictID, "error", err)
		}

		return db.SubCounty{}, fmt.Errorf("upsert subcounty: %w", err)
	}

	// In-app only. Location upserts can happen during imports.
	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_SUBCOUNTY_UPSERTED",
		Title:      "Sub-county updated",
		Severity:   "info",
		Message:    "Surveillance sub-county created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"subcounty_id": item.ID.String(),
			"name":         item.Name,
			"district_id":  item.DistrictID.String(),
		}),
	})

	return item, nil
}

func (s *LocationService) DeleteSubcounty(
	ctx context.Context,
	id uuid.UUID,
) error {
	if s == nil {
		return errors.New("surveillance location service is nil")
	}

	if s.subCountyRepo == nil {
		return errors.New("subcounty repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("subcounty id is required")
	}

	if s.log != nil {
		s.log.Info(ctx, "deleting subcounty", "id", id)
	}

	item, _ := s.subCountyRepo.GetByID(ctx, id)

	if err := s.subCountyRepo.Delete(ctx, id); err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to delete subcounty", "id", id, "error", err)
		}

		return fmt.Errorf("delete subcounty: %w", err)
	}

	if item.ID != uuid.Nil {
		notification := model.Notification{
			Type:       "SURVEILLANCE_SUBCOUNTY_DELETED",
			Title:      "Sub-county deleted",
			Severity:   "warning",
			Message:    "Surveillance sub-county deleted",
			TargetRole: "admin",
			Metadata: utils.MustJSON(map[string]any{
				"subcounty_id": item.ID.String(),
				"name":         item.Name,
				"district_id":  item.DistrictID.String(),
			}),
		}

		s.attachAdminEmailDelivery(
			&notification,
			"admin-alert",
			"Surveillance sub-county deleted",
			fmt.Sprintf("Surveillance sub-county %s was deleted.", item.Name),
			map[string]any{
				"Name":          s.systemAdminName(),
				"Platform":      s.platformName(),
				"Message":       "A surveillance sub-county was deleted.",
				"SubCountyName": item.Name,
				"SubCountyID":   item.ID.String(),
				"DistrictID":    item.DistrictID.String(),
				"ActionURL":     s.adminSurveillanceURL(),
				"Details": fmt.Sprintf(
					"Sub-county ID: %s\nName: %s\nDistrict ID: %s",
					item.ID.String(),
					item.Name,
					item.DistrictID.String(),
				),
			},
		)

		s.notify(ctx, notification)
	}

	return nil
}

func (s *LocationService) notify(
	ctx context.Context,
	notification model.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		if s.log != nil {
			s.log.Error(
				ctx,
				"surveillance location notification failed",
				"type", notification.Type,
				"error", err,
			)
		}
	}
}

func (s *LocationService) attachAdminEmailDelivery(
	notification *model.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminSurveillanceURL()
	}

	notification.Deliveries = []model.NotificationDeliveryRequest{
		{
			Channel: model.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: model.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *LocationService) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *LocationService) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *LocationService) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *LocationService) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *LocationService) adminSurveillanceURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/surveillance"
	}

	return base + "/surveillance"
}
