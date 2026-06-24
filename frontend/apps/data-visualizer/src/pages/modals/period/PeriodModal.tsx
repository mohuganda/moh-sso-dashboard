import {Button, Modal, NumberInput, Select, SelectItem} from "@carbon/react";
import { useMemo, useState } from "react";

import { getAvailablePeriods, getPeriodType, periodType } from "../../Constants.tsx";
import "./period.scss";
import Panel from "../../components/panel/panel.component.tsx";

export default function PeriodModal({ onClose, selected, onSave }) {
  const CURRENT_YEAR = new Date().getFullYear();
  const initialPeriodType = selected?.length > 0 ? getPeriodType(selected[0]?.id) : "Monthly";
  const initialSelectedYear = selected?.length > 0 ? Number(selected[0]?.id?.slice(0,4)) : CURRENT_YEAR;


  const [selectedPeriods, setSelectedPeriods] = useState(selected);
  const initialPeriods = useMemo(
      () => getAvailablePeriods(initialPeriodType, initialSelectedYear).filter(
          (period) => !selectedPeriods?.find((selected) => selected.id === period.id)
      ),
      [initialPeriodType, initialSelectedYear, selectedPeriods],
  );
  const [availablePeriods, setAvailablePeriods] = useState(initialPeriods);
  const [selectedPeriodType, setSelectedPeriodType] = useState(initialPeriodType);
  const [selectedYear, setSelectedYear] = useState(initialSelectedYear);

  const save = () => {
    onSave(selectedPeriods);
    onClose();
  };

  const onChangePeriod = (event) => {
    const paramPeriodType = event?.target?.value;
    setSelectedPeriodType(paramPeriodType);
    const newPeriods = getAvailablePeriods(paramPeriodType, selectedYear)?.filter(
        (period) => !selectedPeriods?.find((selected) => selected?.id === period?.id)
    );
    setAvailablePeriods(newPeriods);
  };

  const onChangeYear = (_event,{ value }) => {
    const paramYear = value;
    setSelectedYear(paramYear);
    const newPeriods = getAvailablePeriods(selectedPeriodType, paramYear)?.filter(
        (period) => !selectedPeriods.find((selected) => selected?.id === period?.id)
    );
    setAvailablePeriods(newPeriods);
  };

  const moveToRight = (selectedItem) => {
    const updatedAvailable = availablePeriods.filter(
        (period) => period?.id !== selectedItem?.id
    );
    setAvailablePeriods(updatedAvailable);
    setSelectedPeriods([...selectedPeriods, selectedItem]);
  };

  const moveToLeft = (selectedItem) => {
    const updatedSelected = selectedPeriods.filter(
        (period) => period?.id !== selectedItem?.id
    );
    setSelectedPeriods(updatedSelected);

    let updatedAvailablePeriods = [...availablePeriods];

    getAvailablePeriods(selectedPeriodType,selectedYear).filter((item) => {
      if (item?.id === selectedItem?.id) {
        updatedAvailablePeriods = [
          ...updatedAvailablePeriods,
          selectedItem,
        ];
      }
    });
    setAvailablePeriods(updatedAvailablePeriods);
  };

  const moveAllToRight = () => {
    if (availablePeriods.length === 0) return;
    setSelectedPeriods([...selectedPeriods, ...availablePeriods]);
    setAvailablePeriods([]);
  };

  const moveAllToLeft = () => {
    if (selectedPeriods.length === 0) return;
    setAvailablePeriods(getAvailablePeriods(selectedPeriodType,selectedYear));
    setSelectedPeriods([]);
  };

  return (
      <Modal
          open
          size="md"
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
          <div className={`panel-container`}>
            <Panel heading={`Reporting Periods`}>
              <ul className={`list`}>
                {availablePeriods.map((period) => (
                    <li
                        role="menuitem"
                        className={`left-list-item`}
                        key={period?.label}
                        onClick={() => moveToRight(period)}
                    >
                      {period?.label}
                    </li>
                ))}
              </ul>
            </Panel>
            <div className={`periods-control-container`}>
              <Button
                  className={`btn-ndwh-mv`}
                  iconDescription="Move all periods to the right"
                  kind="tertiary"
                  hasIconOnly
                  onClick={moveAllToRight}
                  role="button"
                  size="md"
                  disabled={availablePeriods.length < 1}
              >
                <i className="fa-solid fa-angles-right"></i>
              </Button>
              <Button
                  className={`btn-ndwh-mv`}
                  iconDescription="Move all periods to the left"
                  kind="tertiary"
                  hasIconOnly
                  onClick={moveAllToLeft}
                  role="button"
                  size="md"
                  disabled={selectedPeriods.length < 1}
              >
                <i className="fa-solid fa-angles-left"></i>
              </Button>
            </div>
            <Panel heading="Selected Periods">
              <ul className={`list`}>
                {selectedPeriods.map((period) => (
                    <>
                      <li
                          className={`right-list-item`}
                          key={period?.label}
                          role="menuitem"
                          onClick={() => moveToLeft(period)}
                      >
                        {period?.label}
                      </li>
                    </>
                ))}
              </ul>
            </Panel>
          </div>
        </div>
      </Modal>
  );
}