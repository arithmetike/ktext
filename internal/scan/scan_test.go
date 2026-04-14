package scan_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/arithmetike/ktext/internal/scan"
)

func TestScanDepsDevDependenciesOnly(t *testing.T) {
	dir := t.TempDir()

	pkg := map[string]any{
		"name": "@myorg/service",
		"devDependencies": map[string]string{
			"@myorg/shared": "^1.0.0",
			"typescript":    "^5.0.0",
		},
	}
	b, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}

	// Must not panic.
	result := scan.Repo(dir)

	// Should discover the org-scoped sibling from devDependencies.
	found := false
	for _, d := range result.Dependencies {
		if d.Name == "shared" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dependency 'shared' from devDependencies, got %v", result.Dependencies)
	}
}

func TestScanDepsBothDependenciesAndDevDependencies(t *testing.T) {
	dir := t.TempDir()

	pkg := map[string]any{
		"name": "@myorg/app",
		"dependencies": map[string]string{
			"@myorg/core": "^2.0.0",
		},
		"devDependencies": map[string]string{
			"@myorg/testing": "^1.0.0",
		},
	}
	b, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}

	result := scan.Repo(dir)

	names := map[string]bool{}
	for _, d := range result.Dependencies {
		names[d.Name] = true
	}
	if !names["core"] {
		t.Error("expected dependency 'core' from dependencies")
	}
	if !names["testing"] {
		t.Error("expected dependency 'testing' from devDependencies")
	}
}

func TestScanDepsNoDependencies(t *testing.T) {
	dir := t.TempDir()

	pkg := map[string]any{
		"name": "@myorg/empty",
	}
	b, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}

	// Must not panic.
	result := scan.Repo(dir)
	if len(result.Dependencies) != 0 {
		t.Errorf("expected no dependencies, got %v", result.Dependencies)
	}
}
