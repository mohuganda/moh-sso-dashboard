import {Modal, MultiSelect, NumberInput, Select, SelectItem, Tag} from "@carbon/react";
import { useMemo, useState } from "react";

import { getAvailablePeriods, getPeriodType, periodType } from "../../Constants.tsx";
import "./period.css";

export default function PeriodModal({ onClose, selected, onSave }) {
  const CURRENT_YEAR = new Date().getFullYear();
  const initialPeriodType = selected?.length > 0 ? getPeriodType(selected[0]?.id) : "Monthly";
  const initialSelectedYear = selected?.length > 0 ? Number(selected[0]?.id?.slice(0,4)) : CURRENT_YEAR;
  const initialPeriods = useMemo(
    () => getAvailablePeriods(initialPeriodType, initialSelectedYear),
    [initialPeriodType, initialSelectedYear],
  );

  const [availablePeriods, setAvailablePeriods] = useState(initialPeriods);
  const [selectedPeriods, setSelectedPeriods] = useState(selected);
  const [selectedPeriodType, setSelectedPeriodType] = useState(initialPeriodType);
  const [selectedYear, setSelectedYear] = useState(initialSelectedYear);

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

  const onChangeYear = (_event,{ value }) => {
    const paramYear = value;
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

  const handleClearTag = (periodId) => {
    setSelectedPeriods(selectedPeriods.filter(period => period.id !== periodId));
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

        <div className={`multi-select-period-container`}>
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
          <div className={`selected-period-container`}>
            {selectedPeriods.map((period) => (
                <Tag
                    key={period.id}
                    type="blue"
                    filter
                    onClose={() => handleClearTag(period.id)}
                >
                  {period.label}
                </Tag>
            ))}
          </div>
        </div>
      </div>
    </Modal>
  );
}
