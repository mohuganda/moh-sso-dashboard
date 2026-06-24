package admin_units

func toOrgUnitFullResponses(items []OrgUnitFull) []OrgUnitFull {
	if items == nil {
		return []OrgUnitFull{}
	}
	out := make([]OrgUnitFull, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toFacilityResponses(items []Facility) []Facility {
	if items == nil {
		return []Facility{}
	}
	out := make([]Facility, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toOrgUnitSimpleResponses(items []OrgUnitSimple) []OrgUnitSimple {
	if items == nil {
		return []OrgUnitSimple{}
	}
	out := make([]OrgUnitSimple, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toTreeNodeResponses(items []TreeNode) []TreeNode {
	if items == nil {
		return []TreeNode{}
	}
	out := make([]TreeNode, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
