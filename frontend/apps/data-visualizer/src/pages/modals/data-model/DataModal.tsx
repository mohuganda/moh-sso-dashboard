import { Modal, Select, SelectItem } from "@carbon/react";
import { useEffect, useState } from "react";

import {
  type Dataset,
  type ThemeElement,
  useGetDataSetsQuery,
  useLazyGetDataSetElementsQuery,
} from "./data-model.ts";

export default function DataModal({ onClose, selected, onSave }) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedDataset, setSelectedDataset] = useState("");
  const [datasets, setDatasets] = useState<Dataset[] | undefined>([]);
  const [selectedItems, setSelectedItems] = useState(selected);
  const [availableDataSetElements, setAvailableDataSetElements] = useState<ThemeElement[]>([]);
  const [dataSetElementsHolder, setDataSetElementsHolder] = useState<ThemeElement[]>([]);
  const { data: themes, isLoading, error } = useGetDataSetsQuery();
  const [triggerGetTheme] = useLazyGetDataSetElementsQuery();
  // const toast = useToast();

  useEffect(() => {
    if (!isLoading) {
      setDatasets(themes);
    }
    if (error) {
      console.error("Error Encountered while fetching datasets:: " + error);
    }
  }, [error, isLoading, themes]);

  const addItem = (item) => {
    const updatedAvailableParameters = availableDataSetElements.filter(
      (parameter) => parameter !== item,
    );
    setAvailableDataSetElements(updatedAvailableParameters);

    setSelectedItems([...selectedItems, item]);
  };

  const removeItem = (item) => {
    const updatedSelectedItems = selectedItems.filter((parameter) => parameter !== item);
    setSelectedItems(updatedSelectedItems);

    let updatedAvailableElements = [...availableDataSetElements];

    dataSetElementsHolder.filter((parameter) => {
      if (parameter === item) {
        updatedAvailableElements = [...updatedAvailableElements, item];
      }
    });

    setAvailableDataSetElements(updatedAvailableElements);
  };

  const addAll = () => {
    setSelectedItems([...selectedItems, ...filteredAvailableElements]);

    const remainingElements = availableDataSetElements.filter(
      (el) => !filteredAvailableElements.includes(el),
    );
    setAvailableDataSetElements(remainingElements);
    setSearchTerm(""); // Optional: clear search after adding
  };

  const removeAll = () => {
    setAvailableDataSetElements([...dataSetElementsHolder]);
    setSelectedItems([]);
  };

  const save = () => {
    onSave(selectedItems);
    onClose();
  };

  const onChangeSelectedDataSet = async (event) => {
    const theme_id = event?.target?.value;
    setSelectedDataset(theme_id);
    setAvailableDataSetElements([]);
    if (!theme_id) return;

    try {
      const data = await triggerGetTheme(theme_id).unwrap();
      setAvailableDataSetElements(data);
      setDataSetElementsHolder(data);
    } catch (error) {
      setAvailableDataSetElements([]);
      console.error("Error Encountered while fetching data elements:: " + error);
    }
  };

  const filteredAvailableElements = availableDataSetElements?.filter((item) => {
    if (!searchTerm?.trim()) return true;
    return item?.data_element_short_name?.toLowerCase()?.includes(searchTerm?.toLowerCase());
  });

  return (
    <Modal
      open
      size="lg"
      preventCloseOnClickOutside={true}
      hasScrollingContent={true}
      modalHeading="Data"
      secondaryButtonText="Hide"
      primaryButtonText="Update"
      onRequestClose={onClose}
      onRequestSubmit={save}
    >
      <div className="row">
        {/* Left Panel - Available Items */}
        <div className="col-md-6">
          <div className="row mb-3">
            <Select
              id={`data-select`}
              labelText="Dataset"
              defaultValue={selectedDataset}
              onChange={onChangeSelectedDataSet}
            >
              <SelectItem text="" value="" />
              {datasets?.map((dataset: any) => (
                <SelectItem
                  key={dataset?.dataset_id}
                  id={dataset?.dataset_id}
                  value={dataset?.dataset_id}
                  text={dataset?.display_name}
                />
              ))}
            </Select>
          </div>
          {selectedDataset && (
            <div className="mb-3">
              <input
                type="text"
                className="form-control"
                placeholder="Search by dataelement"
                value={searchTerm}
                onChange={(e) => {
                  setSearchTerm(e.target.value);
                }}
              />
            </div>
          )}
        </div>
      </div>
      <div className="row">
        <div className="col-md-6">
          <h6 className="mb-3">Available Data Elements</h6>
          <div className="border" style={{ height: "300px", overflowY: "auto" }}>
            {filteredAvailableElements?.map((item: any) => (
              <div
                key={item?.data_element_key}
                className="p-2 border-bottom d-flex align-items-center"
                style={{ cursor: "pointer" }}
                onClick={() => {
                  addItem(item);
                }}
              >
                <span className="me-2">•</span>
                <div className="small">{item?.data_element_short_name}</div>
              </div>
            ))}
            {availableDataSetElements?.length > 0 && filteredAvailableElements?.length === 0 && (
              <div className="p-3 text-center text-muted">No matches found for "{searchTerm}"</div>
            )}
            {availableDataSetElements?.length === 0 && (
              <div className="p-3 text-center text-muted">
                {selectedDataset ? "All items added or list empty" : "Select a dataset first"}
              </div>
            )}
          </div>
        </div>

        {/* Middle - Transfer Buttons */}
        <div className="col-md-1 d-flex flex-column justify-content-center align-items-center">
          <button
            className="btn btn-outline-primary btn-sm mb-2"
            onClick={addAll}
            disabled={availableDataSetElements?.length === 0}
          >
            <i className="fa-solid fa-angles-right"></i>
          </button>
          <button
            className="btn btn-outline-primary btn-sm"
            onClick={removeAll}
            disabled={selectedItems?.length === 0}
          >
            <i className="fa-solid fa-angles-left"></i>
          </button>
        </div>

        {/* Right Panel - Selected Items */}
        <div className="col-md-5">
          <h6 className="mb-3">Selected Data Elements</h6>
          <div className="border" style={{ height: "300px", overflowY: "auto" }}>
            {selectedItems.length === 0 ? (
              <div className="p-3 text-center text-muted">No items selected</div>
            ) : (
              selectedItems?.map((item) => (
                <div
                  key={item?.data_element_id}
                  className="p-2 border-bottom d-flex align-items-center justify-content-between"
                >
                  <div className="d-flex align-items-center">
                    <span className="me-2">•</span>
                    <div className="small">{item?.data_element_short_name}</div>
                  </div>
                  <button
                    className="btn btn-sm btn-outline-danger"
                    onClick={() => {
                      removeItem(item);
                    }}
                  >
                    <i className="fas fa-times"></i>
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </Modal>
  );
}
