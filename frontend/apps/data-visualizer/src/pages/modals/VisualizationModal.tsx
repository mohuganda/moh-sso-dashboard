import { Modal } from "@carbon/react";

import { chartTypes } from "../Constants";

export default function VisualizationModal({ show, onClose, onChoose, selectedType }) {
  if (!show) return null;

  return (
    <>
      <Modal
        open
        size="md"
        passiveModal
        preventCloseOnClickOutside={true}
        hasScrollingContent={true}
        modalHeading="Select Visualization Type"
        secondaryButtonText="Hide"
        primaryButtonText="Update"
        onRequestClose={onClose}
        onRequestSubmit={onClose}
      >
        <div className="row g-2">
          {chartTypes.map((chart) => {
            const IconComponent = chart?.icon;

            return (
              <div className="col-md-4" key={chart.id || "pivot-table"}>
                <div
                  className={`card h-100 cursor-pointer ${selectedType === chart.id ? "border-info bg-info bg-opacity-10" : ""}`}
                  onClick={() => {
                    onChoose(chart.id);
                    onClose();
                  }}
                  style={{ cursor: "pointer" }}
                >
                  <div className="card-body text-center p-3">
                    <div className="mb-2">
                      {IconComponent && (
                        <IconComponent
                          size={40}
                          className={selectedType === chart.id ? "text-info" : "text-muted"}
                        />
                      )}
                    </div>
                    <h6 className="card-title mb-1 small">{chart.label}</h6>
                    <p className="card-text small text-muted" style={{ fontSize: "0.75rem" }}>
                      {chart.description}
                    </p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </Modal>
    </>
  );
}
