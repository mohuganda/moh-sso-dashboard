import { Checkbox, TreeNode } from "@carbon/react";

export const OrgUnitNode = ({ node, searchTerm, selectedOrgUnit, onSelect, renderRecursive, idPrefix = "check" }) => {
  return (
    <TreeNode
      key={node.id}
      isExpanded={!!searchTerm}
      label={
        <div
          style={{ display: "flex", alignItems: "center", width: "100%" }}
          onClick={(e) => e.stopPropagation()}
        >
          <Checkbox
            id={`${idPrefix}-${node.id}`}
            labelText={node.name}
            checked={selectedOrgUnit === node.name}
            onChange={() => onSelect(node.name)}
          />
        </div>
      }
    >
      {node.children && renderRecursive(node.children, idPrefix)}
    </TreeNode>
  );
};
