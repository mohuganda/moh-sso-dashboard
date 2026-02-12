import { Modal, MultiSelect, NumberInput, Select, SelectItem } from "@carbon/react";
import { useMemo, useState } from "react";

import { getAvailablePeriods, getPeriodType, periodType } from "../Constants";

export default function PeriodModal({ onClose, selected, onSave }) {
  const CURRENT_YEAR = new Date().getFullYear();
  const initialPeriodType = selected?.length > 0 ? getPeriodType(selected[0]?.id) : "Monthly";
  const initialPeriods = useMemo(
    () => getAvailablePeriods(initialPeriodType, CURRENT_YEAR),
    [initialPeriodType, CURRENT_YEAR],
  );

  const [availablePeriods, setAvailablePeriods] = useState(initialPeriods);
  const [selectedPeriods, setSelectedPeriods] = useState(selected);
  const [selectedPeriodType, setSelectedPeriodType] = useState(initialPeriodType);
  const [selectedYear, setSelectedYear] = useState(CURRENT_YEAR);

  const onChangeSelectedPeriod = (event) => {
    setSelectedPeriods(event?.selectedItems);
  };

  const save = () => {
    onSave(selectedPeriods);
    onClose();
  };

  const onChangePeriod = (event) => {
    const paramPeriodType = event?.target?.value;
    setSelectedPeriodType(paramPeriodType);
    const newPeriods = getAvailablePeriods(paramPeriodType, selectedYear);
    setAvailablePeriods(newPeriods);
    setSelectedPeriods([]);
  };

  const onChangeYear = (numberOption) => {
    const paramYear = numberOption?.value;
    setSelectedYear(paramYear);
    const newPeriods = getAvailablePeriods(selectedPeriodType, paramYear);
    setAvailablePeriods(newPeriods);
    setSelectedPeriods([]);
  };

  const comparePeriodItemsItems = (periodA, periodB) => {
    return periodA?.id?.localeCompare(periodB?.id);
  };

  const sortPeriodFunction = (periodItems) => {
    return [...periodItems]?.sort(comparePeriodItemsItems);
  };

  // if (!show) return null;

  return (
    <Modal
      open
      size="sm"
      preventCloseOnClickOutside={true}
      hasScrollingContent={true}
      modalHeading="Period"
      secondaryButtonText="Hide"
      primaryButtonText="Update"
      onRequestClose={onClose}
      onRequestSubmit={save}
    >
      <div className="row pb-3">
        <div className="col-md-12 d-flex">
          <div className="col-md-6 pe-4">
            <Select
              id={`period-type-select`}
              labelText="Period Type"
              onChange={onChangePeriod}
              value={selectedPeriodType ?? ""}
            >
              <SelectItem text="" value="" />
              {periodType?.map((dataset) => (
                <SelectItem key={dataset?.value} value={dataset?.value} text={dataset?.label} />
              ))}
            </Select>
          </div>
          <div className="col-md-6">
            <NumberInput
              defaultValue={CURRENT_YEAR}
              id="period-year-input"
              invalidText="Input is not a valid year"
              label="Year"
              locale="en"
              max={CURRENT_YEAR}
              min={1900}
              size="md"
              step={1}
              type="number"
              value={selectedYear}
              onChange={onChangeYear}
            />
          </div>
        </div>
      </div>
      <div className="row">
        <div className="mb-2 fw-bold">{selectedPeriodType} Periods</div>

        <div style={{ height: "300px", overflowY: "auto" }}>
          <MultiSelect
            id="period-multiselect-id"
            label=""
            titleText="title"
            onChange={onChangeSelectedPeriod}
            hideLabel
            items={availablePeriods}
            sortItems={sortPeriodFunction}
            selectedItems={selectedPeriods}
          />
        </div>
      </div>
    </Modal>
  );
}
