package optimizer

import (
	"os"
	"path/filepath"
	"testing"

	contextmodels "forgeai/backend/internal/context/models"
)

func TestOptimizeSkipsIgnoredDirectoriesAndTruncatesLargeFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "generated"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "generated", "artifact.txt"), []byte("should not be included"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# README\n\nThis is a short readme."), 0o644); err != nil {
		t.Fatal(err)
	}
	largeLines := make([]byte, 0)
	for i := 0; i < 600; i++ {
		largeLines = append(largeLines, []byte("line\n")...)
	}
	if err := os.WriteFile(filepath.Join(dir, "large.go"), largeLines, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := &contextmodels.ExecutionContext{
		Project:       "demo",
		Workspace:     dir,
		CurrentSprint: "Sprint 001",
		Architecture:  "Service-based architecture",
		RelevantFiles: []string{"README.md", "large.go"},
	}

	opt, metrics, err := Optimize(dir, "Implement the feature", "Sprint plan", ctx)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}
	if metrics.OriginalTokens <= 0 {
		t.Fatalf("expected original tokens > 0, got %d", metrics.OriginalTokens)
	}
	if metrics.OptimizedTokens <= 0 {
		t.Fatalf("expected optimized tokens > 0, got %d", metrics.OptimizedTokens)
	}
	if opt == nil {
		t.Fatal("expected optimized context")
	}
	if len(opt.Files) == 0 {
		t.Fatal("expected optimized files to be populated")
	}
	if opt.Files[0].Path == filepath.Join(dir, "generated", "artifact.txt") {
		t.Fatal("expected ignored generated file to be excluded")
	}
	if opt.Files[0].Truncated {
		if opt.Files[0].Summary == "" {
			t.Fatal("expected truncated file to include summary")
		}
	}
}
