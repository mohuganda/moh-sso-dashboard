import { InlineLoading, Modal } from "@carbon/react";
import WeeklyCasesChart from "./surveillance-weekly-cases.component";

type WeeklyCasesPoint = {
  week: string | number;
  value: number;
  label?: string;
};

type FacilityTrendSelection = {
  facilityId?: string;
  facilityName: string;
  regionId?: string;
  districtId?: string;
  subCountyId?: string;
  diseaseId?: string;
  indicatorId?: string;
  diseaseName?: string;
  indicatorName?: string;
};

interface FacilityTrendModalProps {
  open: boolean;
  selectedFacilityTrend: FacilityTrendSelection | null;
  trendSubjectLabel: string;
  selectedFacilityTotalCases: number;
  selectedFacilityPeakWeek: string;
  selectedFacilityTrendData: WeeklyCasesPoint[];
  loading?: boolean;
  onClose: () => void;
}

export function FacilityTrendModal({
  open,
  selectedFacilityTrend,
  trendSubjectLabel,
  selectedFacilityTotalCases,
  selectedFacilityPeakWeek,
  selectedFacilityTrendData,
  loading = false,
  onClose,
}: FacilityTrendModalProps) {
  return (
    <Modal
      open={open}
      modalHeading={
        selectedFacilityTrend
          ? `${selectedFacilityTrend.facilityName} Case Trend`
          : "Facility Case Trend"
      }
      primaryButtonText="Close"
      secondaryButtonText=""
      onRequestClose={onClose}
      onRequestSubmit={onClose}
      size="lg"
    >
      {selectedFacilityTrend ? (
        <div className="disease-details-page__section">
          <dl className="disease-details-page__stats">
            <div>
              <dt>Facility</dt>
              <dd>{selectedFacilityTrend.facilityName}</dd>
            </div>

            <div>
              <dt>Condition</dt>
              <dd>{trendSubjectLabel}</dd>
            </div>

            <div>
              <dt>Total Cases</dt>
              <dd>{loading ? "--" : selectedFacilityTotalCases}</dd>
            </div>

            <div>
              <dt>Peak Week</dt>
              <dd>{loading ? "--" : selectedFacilityPeakWeek}</dd>
            </div>
          </dl>

          {loading ? (
            <InlineLoading description="Loading facility trend..." />
          ) : (
            <WeeklyCasesChart
              diseaseName={`${selectedFacilityTrend.facilityName} - ${trendSubjectLabel}`}
              data={selectedFacilityTrendData}
              height="360px"
            />
          )}
        </div>
      ) : null}
    </Modal>
  );
}
