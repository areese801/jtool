package loganalyzer

import (
	"path/filepath"
	"testing"
)

func TestAnalyzeFile_MultiLine(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "multiline_test.log")
	
	result, err := AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Total lines: %d", result.TotalLines)
	t.Logf("JSON objects: %d", result.JSONLines)
	t.Logf("Skipped lines: %d", result.SkippedLines)
	t.Logf("Unique paths: %d", result.TotalPaths)

	// Should find 7 JSON objects:
	// - 3 user RECORDs (2 multi-line, 1 single-line)
	// - 2 order RECORDs (1 multi-line, 1 single-line)
	// - 2 STATE objects (both multi-line)
	if result.JSONLines != 7 {
		t.Errorf("expected 7 JSON objects, got %d", result.JSONLines)
	}

	// Verify some expected paths exist
	pathsFound := make(map[string]bool)
	for _, p := range result.Paths {
		pathsFound[p.Path] = true
		if p.Path == ".type" {
			t.Logf(".type count=%d, top values:", p.Count)
			for _, v := range p.TopValues {
				t.Logf("  %s (%d)", v.Value, v.Count)
			}
		}
		if p.Path == ".record.email" {
			t.Logf(".record.email distinct=%d, top values:", p.DistinctCount)
			for _, v := range p.TopValues {
				t.Logf("  %s (%d)", v.Value, v.Count)
			}
		}
	}

	expectedPaths := []string{
		".type",
		".stream",
		".record.id",
		".record.name",
		".record.email",
		".record.items[].sku",
		".record.notes",
		".value.users.position",
	}

	for _, p := range expectedPaths {
		if !pathsFound[p] {
			t.Errorf("expected path %s not found", p)
		}
	}
}
