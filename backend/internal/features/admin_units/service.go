package admin_units

import "context"

type Service interface {
	ListOrgUnits(ctx context.Context) ([]OrgUnitFull, error)
	ListFacilities(ctx context.Context) ([]Facility, error)
	ListDistricts(ctx context.Context) ([]OrgUnitSimple, error)
	ListSubCounties(ctx context.Context, subCounty *string) ([]OrgUnitSimple, error)
	ListLocalGovt(ctx context.Context, district *string) ([]OrgUnitSimple, error)
	ListDistrictsByRegion(ctx context.Context, region *string) ([]OrgUnitSimple, error)
	ListRegions(ctx context.Context) ([]OrgUnitSimple, error)
	ListNational(ctx context.Context) ([]OrgUnitSimple, error)
	GetHierarchy(ctx context.Context) ([]TreeNode, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListOrgUnits(ctx context.Context) ([]OrgUnitFull, error) {
	return s.repository.ListOrgUnits(ctx)
}

func (s *service) ListFacilities(ctx context.Context) ([]Facility, error) {
	return s.repository.ListFacilities(ctx)
}

func (s *service) ListDistricts(ctx context.Context) ([]OrgUnitSimple, error) {
	return s.repository.ListDistricts(ctx)
}

func (s *service) ListSubCounties(ctx context.Context, subCounty *string) ([]OrgUnitSimple, error) {
	return s.repository.ListSubCounties(ctx, subCounty)
}

func (s *service) ListLocalGovt(ctx context.Context, district *string) ([]OrgUnitSimple, error) {
	return s.repository.ListLocalGovt(ctx, district)
}

func (s *service) ListDistrictsByRegion(ctx context.Context, region *string) ([]OrgUnitSimple, error) {
	return s.repository.ListDistrictsByRegion(ctx, region)
}

func (s *service) ListRegions(ctx context.Context) ([]OrgUnitSimple, error) {
	return s.repository.ListRegions(ctx)
}

func (s *service) ListNational(ctx context.Context) ([]OrgUnitSimple, error) {
	return s.repository.ListNational(ctx)
}

func (s *service) GetHierarchy(ctx context.Context) ([]TreeNode, error) {
	data, err := s.repository.ListHierarchy(ctx)
	if err != nil {
		return nil, err
	}
	return buildHierarchyTree(data), nil
}
