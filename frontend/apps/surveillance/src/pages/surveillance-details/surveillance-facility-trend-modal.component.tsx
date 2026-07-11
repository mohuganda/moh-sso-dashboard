import { useEffect } from "react";

import { InlineLoading } from "@carbon/react";

import { useModal } from "@moh-sso/ui";

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
  const { openModal, closeModal } = useModal();

  useEffect(() => {
    if (!open) {
      closeModal();
      return;
    }

    const closeTrendModal = () => {
      closeModal();
      onClose();
    };

    openModal({
      title: selectedFacilityTrend
        ? `${selectedFacilityTrend.facilityName} Case Trend`
        : "Facility Case Trend",
      onClose: onClose,
      size: "lg",
      content: selectedFacilityTrend ? (
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
      ) : null,
      primaryAction: {
        label: "Close",
        onClick: closeTrendModal,
      },
    });
  }, [
    closeModal,
    loading,
    onClose,
    open,
    openModal,
    selectedFacilityPeakWeek,
    selectedFacilityTotalCases,
    selectedFacilityTrend,
    selectedFacilityTrendData,
    trendSubjectLabel,
  ]);

  return null;
}
