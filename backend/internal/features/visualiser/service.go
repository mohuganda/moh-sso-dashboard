package visualiser

import "context"

type Service interface {
	ListDatasets(ctx context.Context) ([]DatasetResponse, error)
	ListDataElements(ctx context.Context, dataSetID *string) ([]DataElementResponse, error)
	ListDataValues(ctx context.Context, req DataValuesRequest) (DataValuesResponse, error)
	ListThemes(ctx context.Context) ([]ThemeResponse, error)
	ListDataElementsByTheme(ctx context.Context, themeID string) ([]DataElementByThemeResponse, error)
	ListHIVSummary(ctx context.Context) ([]HIVSummaryResponse, error)
	ListHIVTested(ctx context.Context) ([]HIVTestedResponse, error)
	ListHIVRegimen(ctx context.Context) ([]HIVRegimenResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListDatasets(ctx context.Context) ([]DatasetResponse, error) {
	return s.repository.ListDatasets(ctx)
}

func (s *service) ListDataElements(ctx context.Context, dataSetID *string) ([]DataElementResponse, error) {
	return s.repository.ListDataElements(ctx, dataSetID)
}

func (s *service) ListDataValues(ctx context.Context, req DataValuesRequest) (DataValuesResponse, error) {
	return s.repository.ListDataValues(ctx, req)
}

func (s *service) ListThemes(ctx context.Context) ([]ThemeResponse, error) {
	return s.repository.ListThemes(ctx)
}

func (s *service) ListDataElementsByTheme(ctx context.Context, themeID string) ([]DataElementByThemeResponse, error) {
	return s.repository.ListDataElementsByTheme(ctx, themeID)
}

func (s *service) ListHIVSummary(ctx context.Context) ([]HIVSummaryResponse, error) {
	return s.repository.ListHIVSummary(ctx)
}

func (s *service) ListHIVTested(ctx context.Context) ([]HIVTestedResponse, error) {
	return s.repository.ListHIVTested(ctx)
}

func (s *service) ListHIVRegimen(ctx context.Context) ([]HIVRegimenResponse, error) {
	return s.repository.ListHIVRegimen(ctx)
}
