package paths

import (
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name           string
		input          any
		expectedPaths  int
		expectedLeafs  int
		checkPath      string // Optional: verify this path exists
		checkPathCount int    // Expected count for checkPath
	}{
		{
			name:          "simple object",
			input:         map[string]any{"name": "Alice", "age": 30},
			expectedPaths: 2,
			expectedLeafs: 2,
			checkPath:     ".name",
			checkPathCount: 1,
		},
		{
			name:          "nested object",
			input:         map[string]any{"user": map[string]any{"name": "Bob"}},
			expectedPaths: 1,
			expectedLeafs: 1,
			checkPath:     ".user.name",
			checkPathCount: 1,
		},
		{
			name: "array of objects",
			input: map[string]any{
				"users": []any{
					map[string]any{"name": "Alice"},
					map[string]any{"name": "Bob"},
				},
			},
			expectedPaths:  1, // .users[].name appears once (unique path)
			expectedLeafs:  2, // but with count=2
			checkPath:      ".users[].name",
			checkPathCount: 2,
		},
		{
			name: "mixed types",
			input: map[string]any{
				"string": "hello",
				"number": 42,
				"bool":   true,
				"null":   nil,
			},
			expectedPaths: 4,
			expectedLeafs: 4,
		},
		{
			name:          "empty object",
			input:         map[string]any{},
			expectedPaths: 0,
			expectedLeafs: 0,
		},
		{
			name:           "simple array",
			input:          []any{1, 2, 3},
			expectedPaths:  1,  // []
			expectedLeafs:  3,  // count=3
			checkPath:      "[]",
			checkPathCount: 3,
		},
		{
			name: "deeply nested",
			input: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": "deep",
						},
					},
				},
			},
			expectedPaths:  1,
			expectedLeafs:  1,
			checkPath:      ".a.b.c.d",
			checkPathCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Extract(tt.input)

			if result.TotalPaths != tt.expectedPaths {
				t.Errorf("TotalPaths = %d, want %d", result.TotalPaths, tt.expectedPaths)
			}

			if result.TotalLeafs != tt.expectedLeafs {
				t.Errorf("TotalLeafs = %d, want %d", result.TotalLeafs, tt.expectedLeafs)
			}

			// Check specific path if requested
			if tt.checkPath != "" {
				found := false
				for _, p := range result.Paths {
					if p.Path == tt.checkPath {
						found = true
						if p.Count != tt.checkPathCount {
							t.Errorf("path %q count = %d, want %d", tt.checkPath, p.Count, tt.checkPathCount)
						}
						break
					}
				}
				if !found {
					t.Errorf("path %q not found in result", tt.checkPath)
				}
			}
		})
	}
}

func TestExtractSorted(t *testing.T) {
	// Verify paths are sorted alphabetically
	input := map[string]any{
		"zebra": 1,
		"apple": 2,
		"mango": 3,
	}

	result := Extract(input)

	if len(result.Paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(result.Paths))
	}

	expected := []string{".apple", ".mango", ".zebra"}
	for i, p := range result.Paths {
		if p.Path != expected[i] {
			t.Errorf("path[%d] = %q, want %q", i, p.Path, expected[i])
		}
	}
}
