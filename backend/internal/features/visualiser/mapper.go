package visualiser

func toDatasetResponses(items []DatasetResponse) []DatasetResponse {
	if items == nil {
		return []DatasetResponse{}
	}
	out := make([]DatasetResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toDataElementResponses(items []DataElementResponse) []DataElementResponse {
	if items == nil {
		return []DataElementResponse{}
	}
	out := make([]DataElementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toDataValuesResponse(item DataValuesResponse) DataValuesResponse {
	if item.Rows == nil {
		item.Rows = []DataValueRowResponse{}
	}
	return item
}

func toThemeResponses(items []ThemeResponse) []ThemeResponse {
	if items == nil {
		return []ThemeResponse{}
	}
	out := make([]ThemeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toDataElementByThemeResponses(items []DataElementByThemeResponse) []DataElementByThemeResponse {
	if items == nil {
		return []DataElementByThemeResponse{}
	}
	out := make([]DataElementByThemeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toHIVSummaryResponses(items []HIVSummaryResponse) []HIVSummaryResponse {
	if items == nil {
		return []HIVSummaryResponse{}
	}
	out := make([]HIVSummaryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toHIVTestedResponses(items []HIVTestedResponse) []HIVTestedResponse {
	if items == nil {
		return []HIVTestedResponse{}
	}
	out := make([]HIVTestedResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toHIVRegimenResponses(items []HIVRegimenResponse) []HIVRegimenResponse {
	if items == nil {
		return []HIVRegimenResponse{}
	}
	out := make([]HIVRegimenResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
