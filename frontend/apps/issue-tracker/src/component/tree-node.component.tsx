import { Checkbox, TreeNode } from "@carbon/react";

export const OrgUnitNode = ({ node, searchTerm, selectedOrgUnit, onSelect, renderRecursive }) => {
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
            id={`check-${node.id}`}
            labelText={node.name}
            checked={selectedOrgUnit === node.name}
            onChange={() => onSelect(node.name)}
          />
        </div>
      }
    >
      {node.children && renderRecursive(node.children)}
    </TreeNode>
  );
};
