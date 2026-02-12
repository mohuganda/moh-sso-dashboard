import { Modal, Select, SelectItem } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import API from "../../helpers/api";

type DataElement = {
  id: string;
  name: string;
};

type ElementsMap = Record<string, DataElement[]>;

export default function DataModal({ onClose, selected, onSave }) {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedDataset] = useState("");
  const [datasets, setDatasets] = useState([]);
  const [datasetToElements] = useState({});
  const [selectedItems, setSelectedItems] = useState(selected);
  const [availableDataSetElements, setAvailableDataSetElements] = useState([]);
  // const toast = useToast();

  const fetchData = async () => {
    const { status, data } = await API.get("/visualizer/themes");
    if (status === 200) {
      setDatasets(data);
    } else {
      //toast.error("Error Encountered while fetching datasets");
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const filteredItems = useMemo(() => {
    // 1. Ensure a valid map object exists
    const elementsMap: ElementsMap = datasetToElements ?? {};

    let elements;

    if (selectedDataset) {
      elements = elementsMap[selectedDataset] ?? [];
    } else {
      elements = Object.entries(elementsMap).flatMap(([ds, list]) =>
        list.map((de) => ({ ds, de })),
      );
    }

    // 3. Normalize to common shape { ds, de }
    const normalized = (elements || []).map((el) => {
      // If element is a string, assume it's the data element from the selectedDataset context
      if (typeof el === "string") {
        return { ds: selectedDataset || "", de: el };
      }
      // If it's already an object {ds, de}, return it
      return el;
    });

    // 4. Map to final output shape { id, name, dataset }
    const items = normalized.map((row) => {
      const dataElement = String(row.de || "");
      const datasetName = row.ds || selectedDataset || "";

      return { id: dataElement, name: dataElement, dataset: datasetName };
    });

    // 5. Apply Search Filter
    const term = (searchTerm || "").toLowerCase();
    const searched = term
      ? items.filter((it) => (it.id || "").toLowerCase().includes(term))
      : items;

    // 6. Sort
    searched.sort(
      (a, b) =>
        (a.dataset || "").localeCompare(b.dataset || "") || (a.id || "").localeCompare(b.id || ""),
    );

    return searched;
  }, [datasetToElements, selectedDataset, searchTerm]);

  const addItem = (item) => {
    if (!selectedItems?.find((selected) => selected?.data_element_id === item.data_element_id)) {
      setSelectedItems([...selectedItems, item]);
    }
  };

  const removeItem = (item) => {
    setSelectedItems(
      selectedItems?.filter((selected) => selected?.data_element_id !== item?.data_element_id),
    );
  };

  const addAll = () => {
    const newItems = filteredItems?.filter(
      (item) =>
        !selectedItems?.find((selected) => selected?.data_element_id === item?.data_element_id),
    );
    setSelectedItems([...selectedItems, ...newItems]);
  };

  const removeAll = () => {
    setSelectedItems([]);
  };

  const save = () => {
    onSave(selectedItems);
    onClose();
  };

  const onChangeSelectedDataSet = async (event) => {
    const theme_id = event?.target?.value;
    const { status, data } = await API.post("/visualizer/dataelements/theme", {
      theme_id: theme_id,
    });
    if (status === 200) {
      setAvailableDataSetElements(data);
    } else {
      // toast.error("Error Encountered while fetching data elements");
    }
  };

  // if (!show) return null;

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

          <div className="row mb-3">
            <Select
              id={`data-select`}
              labelText="Dataset"
              defaultValue={selectedDataset}
              onChange={onChangeSelectedDataSet}
            >
              <SelectItem text="" value="" />
              {datasets?.map((dataset: any) => (
                <SelectItem value={dataset?.theme_id} text={dataset?.theme_name} />
              ))}
            </Select>
          </div>
        </div>
      </div>
      <div className="row">
        <div className="col-md-6">
          <h6 className="mb-3">Available Data Elements</h6>
          <div className="border" style={{ height: "300px", overflowY: "auto" }}>
            {availableDataSetElements?.map((item: any) => (
              <div
                key={item?.data_element_id}
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
            {availableDataSetElements?.length === 0 && (
              <div className="p-3 text-center text-muted">No items found</div>
            )}
          </div>
        </div>

        {/* Middle - Transfer Buttons */}
        <div className="col-md-1 d-flex flex-column justify-content-center align-items-center">
          <button
            className="btn btn-outline-primary btn-sm mb-2"
            onClick={addAll}
            disabled={filteredItems.length === 0}
          >
            <i className="fas fa-arrow-right"></i>
          </button>
          <button
            className="btn btn-primary btn-sm mb-2"
            onClick={() => {
              // Add first selected item from filtered list
              const firstItem = filteredItems[0];
              if (firstItem) addItem(firstItem);
            }}
            disabled={filteredItems?.length === 0}
          >
            <i className="fas fa-arrow-right"></i>
          </button>
          <button
            className="btn btn-primary btn-sm mb-2"
            onClick={() => {
              // Remove first selected item
              const firstSelected = selectedItems[0];
              if (firstSelected) removeItem(firstSelected);
            }}
            disabled={selectedItems?.length === 0}
          >
            <i className="fas fa-arrow-left"></i>
          </button>
          <button
            className="btn btn-outline-primary btn-sm"
            onClick={removeAll}
            disabled={selectedItems?.length === 0}
          >
            <i className="fas fa-arrow-left"></i>
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
