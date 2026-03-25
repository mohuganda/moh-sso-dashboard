import { Modal } from "@carbon/react";
import { useEffect, useState } from "react";
import { useGetHierarchyQuery } from "./org-unit.ts";

function TreeNode({ node, level = 0, selectedUnits, onToggle, onExpand, expandedNodes }) {
  const childrenObj = node.children || {};

  const uniqueChildren = Object.values(childrenObj).filter(
    (child: any, index, self) => index === self.findIndex((c: any) => c.id === child.id),
  );

  const hasChildren = uniqueChildren.length > 0;
  const isExpanded = expandedNodes.has(node.uid);
  const isSelected = selectedUnits.has(node.uid);
  const childCount = uniqueChildren.length;

  const handleToggle = () => {
    onToggle(node.uid);
  };

  const handleExpand = () => {
    if (hasChildren) {
      onExpand(node.uid);
    }
  };

  const getIcon = () => {
    if (hasChildren) {
      return isExpanded ? "fas fa-chevron-down" : "fas fa-chevron-right";
    }
    return "";
  };

  const indent = level * 18;
  const childrenPadding = indent + 18;

  return (
    <div>
      <div
        className="d-flex align-items-center p-1"
        style={{
          paddingLeft: `${indent}px`,
        }}
      >
        <button
          className="btn btn-sm btn-link p-0 me-2"
          style={{ width: "16px", visibility: hasChildren ? "visible" : "hidden" }}
          onClick={handleExpand}
          disabled={!hasChildren}
        >
          {hasChildren && <i className={getIcon()}></i>}
        </button>
        <input
          type="checkbox"
          className="form-check-input me-2"
          checked={isSelected}
          onChange={handleToggle}
        />
        <span className="d-flex">
          {node.name}
          {childCount > 0 && <span className="text-muted"> ({childCount})</span>}
        </span>
      </div>

      {hasChildren && isExpanded && (
        <div
          style={{
            paddingLeft: `${childrenPadding}px`,
            borderLeft: "1px solid #eee",
          }}
        >
          {uniqueChildren.map((child: any) => (
            <TreeNode
              key={child.uid}
              node={child}
              level={level + 1}
              selectedUnits={selectedUnits}
              onToggle={onToggle}
              onExpand={onExpand}
              expandedNodes={expandedNodes}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default function OrgUnitModal({ onClose, selected, onSave }) {
  const [selectedUnits, setSelectedUnits] = useState(new Set(selected));
  const [expandedNodes, setExpandedNodes] = useState(new Set());
  const [orgUnits, setOrgUnits] = useState<any>({});
  const { data: hierarchyData, isLoading, error} = useGetHierarchyQuery()

  useEffect(() => {
    if(!isLoading) {
      console.log(hierarchyData);

      setOrgUnits(hierarchyData);

      const expandIds = new Set();
      const rootKeys = Object.keys(hierarchyData);

      if (rootKeys.length > 0) {
        const firstRootNode = hierarchyData[rootKeys[0]];
        expandIds.add(firstRootNode.uid);

        const firstChildKeys = Object.keys(firstRootNode.children || {});
        if (firstChildKeys.length > 0) {
          const firstChildNode = firstRootNode.children[firstChildKeys[0]];
          expandIds.add(firstChildNode.uid);
        }
      }
      setExpandedNodes(expandIds);
    }
  }, [hierarchyData, isLoading]);


  useEffect(() => {
    setSelectedUnits(new Set(selected));
  }, [selected]);

  const toggleUnit = (unitUid) => {
    const newSelected = new Set(selectedUnits);
    if (newSelected.has(unitUid)) {
      newSelected.delete(unitUid);
    } else {
      newSelected.add(unitUid);
    }
    setSelectedUnits(newSelected);
  };

  const toggleExpanded = (unitUid) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(unitUid)) {
      newExpanded.delete(unitUid);
    } else {
      newExpanded.add(unitUid);
    }
    setExpandedNodes(newExpanded);
  };

  const deselectAll = () => {
    setSelectedUnits(new Set());
  };

  const save = () => {
    onSave(Array.from(selectedUnits));
    onClose();
  };

  // if (!show) return null;

  const selectedCount = selectedUnits.size;

  return (
    <>
      <Modal
        open
        size="md"
        preventCloseOnClickOutside={true}
        hasScrollingContent={true}
        modalHeading="Organisation Units"
        secondaryButtonText="Hide"
        primaryButtonText="Update"
        onRequestClose={onClose}
        onRequestSubmit={save}
      >
        <div className="border" style={{ height: "400px", overflowY: "auto" }}>
          {isLoading ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="spinner-border text-primary" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
          ) : error ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="text-center text-danger">
                <i className="fas fa-exclamation-triangle fa-2x mb-2"></i>
                <div>
                  {'status' in error
                      ? `Error ${error.status}: ${JSON.stringify(error.data)}`
                      : error.message}
                </div>
              </div>
            </div>
          ) : Object.keys(orgUnits).length === 0 ? (
            <div className="d-flex justify-content-center align-items-center h-100">
              <div className="text-center text-muted">
                <i className="fas fa-folder-open fa-2x mb-2"></i>
                <div>No organizational units found</div>
              </div>
            </div>
          ) : (
            Object.values(orgUnits).map((unit: any) => (
              <TreeNode
                key={unit.uid}
                node={unit}
                selectedUnits={selectedUnits}
                onToggle={toggleUnit}
                onExpand={toggleExpanded}
                expandedNodes={expandedNodes}
              />
            ))
          )}
        </div>

        <div className="mt-3 d-flex justify-content-between align-items-center">
          <span className="text-muted">
            {selectedCount} selected
            {selectedCount > 0 && (
              <button className="btn btn-link p-0 ms-2 text-decoration-none" onClick={deselectAll}>
                - Deselect all
              </button>
            )}
          </span>
        </div>
      </Modal>
    </>
  );
}
