package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const modulePath = "github.com/moh-sso-dashboard/internal/features/"

var allowedFeatureImports = map[string]map[string]string{
	"announcements": {
		"users": "announcement audiences can resolve Keycloak users through the users repository port",
		"rbac":  "announcement group audiences resolve members through the RBAC repository port",
	},
	"auth": {
		"authsession": "auth owns the authsession backing store as an implementation detail",
	},
	"documents": {
		"storage_locations": "document upload flows validate configured storage locations",
	},
	"email": {
		"rbac":  "email group targeting resolves recipients through the RBAC repository port",
		"users": "email group targeting filters resolved members through the users repository port",
	},
	"rbac": {
		"system_rbac": "RBAC sync applies the system RBAC seed format",
	},
}

func TestFeaturePackagesDoNotImportOtherFeaturesWithoutApproval(t *testing.T) {
	featuresDir := filepath.Join(repoRoot(t), "internal", "features")

	err := filepath.WalkDir(featuresDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		sourceFeature, ok := featureName(featuresDir, path)
		if !ok {
			return nil
		}

		fileSet := token.NewFileSet()
		file, parseErr := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}

		for _, importSpec := range file.Imports {
			targetFeature, ok := importedFeature(importSpec)
			if !ok || targetFeature == sourceFeature {
				continue
			}
			if reason := allowedReason(sourceFeature, targetFeature); reason == "" {
				relPath, _ := filepath.Rel(repoRoot(t), path)
				t.Errorf("%s imports feature %q from feature %q without an explicit boundary exception", relPath, targetFeature, sourceFeature)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func featureName(featuresDir, path string) (string, bool) {
	relPath, err := filepath.Rel(featuresDir, path)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", false
	}
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) < 2 {
		return "", false
	}
	return parts[0], true
}

func importedFeature(importSpec *ast.ImportSpec) (string, bool) {
	importPath := strings.Trim(importSpec.Path.Value, `"`)
	if !strings.HasPrefix(importPath, modulePath) {
		return "", false
	}
	remainder := strings.TrimPrefix(importPath, modulePath)
	parts := strings.Split(remainder, "/")
	if parts[0] == "" {
		return "", false
	}
	return parts[0], true
}

func allowedReason(sourceFeature, targetFeature string) string {
	if targets, ok := allowedFeatureImports[sourceFeature]; ok {
		return targets[targetFeature]
	}
	return ""
}
