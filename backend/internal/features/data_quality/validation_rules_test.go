package data_quality

import "testing"

func TestNormalizeValidationRuleRequest(t *testing.T) {
	value := "4"
	tableID := "cht_form_097b_with_vhts"
	category := "custom"
	description := "Number of ANC visits should be at least 4 or more"

	input, message := normalizeValidationRuleRequest(validationRuleImportRequest{
		TableID:     &tableID,
		Category:    &category,
		Code:        "MAL_900",
		Severity:    "ERROR",
		Description: &description,
		Column:      "atleast_4_anc_visits",
		Op:          "GTE",
		Value:       &value,
	}, "user-1")

	if message != "" {
		t.Fatalf("expected valid rule, got validation message %q", message)
	}
	if input.Code != "MAL_900" {
		t.Fatalf("expected code to be preserved, got %q", input.Code)
	}
	if input.Severity != "error" {
		t.Fatalf("expected normalized severity, got %q", input.Severity)
	}
	if input.Op != "gte" {
		t.Fatalf("expected normalized operator, got %q", input.Op)
	}
}

func TestNormalizeValidationRuleRequestAllowsProgramWithoutCategory(t *testing.T) {
	program := "malaria"
	valueColumn := "atleast_8_anc_visits"

	_, message := normalizeValidationRuleRequest(validationRuleImportRequest{
		Program:     &program,
		Code:        "67890",
		Severity:    "error",
		Column:      "atleast_8_anc_visits",
		Op:          "gte",
		ValueColumn: &valueColumn,
	}, nil)

	if message != "" {
		t.Fatalf("expected program-only rule to be valid, got %q", message)
	}
}

func TestNormalizeValidationRuleRequestRejectsMissingComparisonTarget(t *testing.T) {
	_, message := normalizeValidationRuleRequest(validationRuleImportRequest{
		Code:     "MAL_900",
		Severity: "error",
		Column:   "atleast_4_anc_visits",
		Op:       "gte",
	}, nil)

	if message == "" {
		t.Fatal("expected validation message")
	}
}

func TestNormalizeValidationRuleRequestAllowsNullOperatorsWithoutTarget(t *testing.T) {
	_, message := normalizeValidationRuleRequest(validationRuleImportRequest{
		Code:     "NULL_1",
		Severity: "warning",
		Column:   "submitted_at",
		Op:       "notnull",
	}, nil)

	if message != "" {
		t.Fatalf("expected notnull rule to be valid, got %q", message)
	}
}
