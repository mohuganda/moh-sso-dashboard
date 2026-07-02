package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/moh-sso-dashboard/internal/version"
)

func TestWriteVersionText(t *testing.T) {
	var output bytes.Buffer
	if err := writeVersion(&output, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "moh-sso-dashboard-backend") {
		t.Fatalf("unexpected output: %s", output.String())
	}
}

func TestWriteVersionJSON(t *testing.T) {
	var output bytes.Buffer
	if err := writeVersion(&output, "json"); err != nil {
		t.Fatal(err)
	}
	var info version.BuildInfo
	if err := json.Unmarshal(output.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Service != "moh-sso-dashboard-backend" {
		t.Fatalf("unexpected service: %s", info.Service)
	}
}

func TestWriteVersionRejectsUnknownOutput(t *testing.T) {
	if err := writeVersion(&bytes.Buffer{}, "yaml"); err == nil {
		t.Fatal("expected unsupported output error")
	}
}
