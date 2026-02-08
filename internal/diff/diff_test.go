package diff

import (
	"encoding/json"
	"testing"

	"jtool/internal/normalize"
)

// TestCompare uses table-driven tests, a common Go testing pattern.
func TestCompare(t *testing.T) {
	tests := []struct {
		name          string
		leftJSON      string
		rightJSON     string
		expectedType  DiffType
		expectedStats DiffStats
	}{
		{
			name:          "equal primitives - string",
			leftJSON:      `"hello"`,
			rightJSON:     `"hello"`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1},
		},
		{
			name:          "equal primitives - number",
			leftJSON:      `42`,
			rightJSON:     `42`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1},
		},
		{
			name:          "equal primitives - boolean",
			leftJSON:      `true`,
			rightJSON:     `true`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1},
		},
		{
			name:          "equal primitives - null",
			leftJSON:      `null`,
			rightJSON:     `null`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1},
		},
		{
			name:          "changed primitive - string",
			leftJSON:      `"hello"`,
			rightJSON:     `"world"`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Changed: 1},
		},
		{
			name:          "changed primitive - number",
			leftJSON:      `42`,
			rightJSON:     `99`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Changed: 1},
		},
		{
			name:          "type change - string to number",
			leftJSON:      `"42"`,
			rightJSON:     `42`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Changed: 1},
		},
		{
			name:          "added - null to value",
			leftJSON:      `null`,
			rightJSON:     `"hello"`,
			expectedType:  DiffAdded,
			expectedStats: DiffStats{Added: 1},
		},
		{
			name:          "removed - value to null",
			leftJSON:      `"hello"`,
			rightJSON:     `null`,
			expectedType:  DiffRemoved,
			expectedStats: DiffStats{Removed: 1},
		},
		{
			name:          "equal empty objects",
			leftJSON:      `{}`,
			rightJSON:     `{}`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1}, // Empty object is itself a leaf
		},
		{
			name:          "equal simple objects",
			leftJSON:      `{"a": 1, "b": 2}`,
			rightJSON:     `{"a": 1, "b": 2}`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 2},
		},
		{
			name:          "object with added key",
			leftJSON:      `{"a": 1}`,
			rightJSON:     `{"a": 1, "b": 2}`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 1, Added: 1},
		},
		{
			name:          "object with removed key",
			leftJSON:      `{"a": 1, "b": 2}`,
			rightJSON:     `{"a": 1}`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 1, Removed: 1},
		},
		{
			name:          "object with changed value",
			leftJSON:      `{"a": 1, "b": 2}`,
			rightJSON:     `{"a": 1, "b": 999}`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 1, Changed: 1},
		},
		{
			name:          "equal empty arrays",
			leftJSON:      `[]`,
			rightJSON:     `[]`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 1}, // Empty array is itself a leaf
		},
		{
			name:          "equal simple arrays",
			leftJSON:      `[1, 2, 3]`,
			rightJSON:     `[1, 2, 3]`,
			expectedType:  DiffEqual,
			expectedStats: DiffStats{Equal: 3},
		},
		{
			name:          "array with added element",
			leftJSON:      `[1, 2]`,
			rightJSON:     `[1, 2, 3]`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 2, Added: 1},
		},
		{
			name:          "array with removed element",
			leftJSON:      `[1, 2, 3]`,
			rightJSON:     `[1, 2]`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 2, Removed: 1},
		},
		{
			name:          "array with changed element",
			leftJSON:      `[1, 2, 3]`,
			rightJSON:     `[1, 999, 3]`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 2, Changed: 1},
		},
		{
			name:          "nested object change",
			leftJSON:      `{"user": {"name": "Alice", "age": 30}}`,
			rightJSON:     `{"user": {"name": "Bob", "age": 30}}`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 1, Changed: 1},
		},
		{
			name:          "deeply nested change",
			leftJSON:      `{"a": {"b": {"c": {"d": 1}}}}`,
			rightJSON:     `{"a": {"b": {"c": {"d": 2}}}}`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Changed: 1},
		},
		{
			name:          "array of objects",
			leftJSON:      `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`,
			rightJSON:     `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Charlie"}]`,
			expectedType:  DiffChanged,
			expectedStats: DiffStats{Equal: 3, Changed: 1},
		},
	}

	for _, tt := range tests {
		// t.Run creates a subtest, like pytest's parametrize
		// You can run specific tests with: go test -run "TestCompare/equal_primitives"
		t.Run(tt.name, func(t *testing.T) {
			// Parse JSON
			var left, right any
			if err := json.Unmarshal([]byte(tt.leftJSON), &left); err != nil {
				t.Fatalf("failed to parse leftJSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.rightJSON), &right); err != nil {
				t.Fatalf("failed to parse rightJSON: %v", err)
			}

			// Run the comparison
			result := Compare(left, right)

			// Check root type
			if result.Root.Type != tt.expectedType {
				t.Errorf("expected root type %q, got %q", tt.expectedType, result.Root.Type)
			}

			// Check stats
			if result.Stats != tt.expectedStats {
				t.Errorf("expected stats %+v, got %+v", tt.expectedStats, result.Stats)
			}
		})
	}
}

// TestComparePaths verifies that JSON paths are correctly tracked
func TestComparePaths(t *testing.T) {
	leftJSON := `{"users": [{"name": "Alice"}, {"name": "Bob"}]}`
	rightJSON := `{"users": [{"name": "Alice"}, {"name": "Charlie"}]}`

	var left, right any
	json.Unmarshal([]byte(leftJSON), &left)
	json.Unmarshal([]byte(rightJSON), &right)

	result := Compare(left, right)

	// Find the changed node
	var changedPath string
	var findChanged func(node DiffNode)
	findChanged = func(node DiffNode) {
		if node.Type == DiffChanged && len(node.Children) == 0 {
			changedPath = node.Path
			return
		}
		for _, child := range node.Children {
			findChanged(child)
		}
	}
	findChanged(result.Root)

	expectedPath := ".users[1].name"
	if changedPath != expectedPath {
		t.Errorf("expected changed path %q, got %q", expectedPath, changedPath)
	}
}

// TestCompareNilInputs verifies handling of nil inputs
func TestCompareNilInputs(t *testing.T) {
	tests := []struct {
		name         string
		left         any
		right        any
		expectedType DiffType
	}{
		{
			name:         "both nil",
			left:         nil,
			right:        nil,
			expectedType: DiffEqual,
		},
		{
			name:         "left nil",
			left:         nil,
			right:        "hello",
			expectedType: DiffAdded,
		},
		{
			name:         "right nil",
			left:         "hello",
			right:        nil,
			expectedType: DiffRemoved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.left, tt.right)
			if result.Root.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, result.Root.Type)
			}
		})
	}
}

// Example test showing Go's example-based documentation
// These appear in godoc and can be run as tests
func ExampleCompare() {
	left := map[string]any{"name": "Alice", "age": 30.0}
	right := map[string]any{"name": "Bob", "age": 30.0}

	result := Compare(left, right)

	// Output will vary, so we just demonstrate usage
	_ = result.Stats.Changed // Would be 1
}

// ============================================================
// QA Tests for Checkbox Options with CompareWithOptions
// These test the integration of normalization with diff
// ============================================================

// TestCompareWithOptionsSortKeys tests the Sort Keys checkbox
func TestCompareWithOptionsSortKeys(t *testing.T) {
	tests := []struct {
		name         string
		leftJSON     string
		rightJSON    string
		sortKeys     bool
		expectEqual  bool
		expectedType DiffType
	}{
		{
			name:         "same keys different order - sort enabled - should be equal",
			leftJSON:     `{"a": 1, "b": 2}`,
			rightJSON:    `{"b": 2, "a": 1}`,
			sortKeys:     true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "nested same keys different order - sort enabled",
			leftJSON:     `{"outer": {"x": 1, "a": 2}}`,
			rightJSON:    `{"outer": {"a": 2, "x": 1}}`,
			sortKeys:     true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "array of objects - keys reordered",
			leftJSON:     `[{"z": 1, "a": 2}]`,
			rightJSON:    `[{"a": 2, "z": 1}]`,
			sortKeys:     true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "actual difference - not just order",
			leftJSON:     `{"a": 1, "b": 2}`,
			rightJSON:    `{"a": 1, "b": 999}`,
			sortKeys:     true,
			expectEqual:  false,
			expectedType: DiffChanged,
		},
		{
			name:         "different keys - not equal",
			leftJSON:     `{"a": 1}`,
			rightJSON:    `{"b": 1}`,
			sortKeys:     true,
			expectEqual:  false,
			expectedType: DiffChanged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			opts := normalize.Options{SortKeys: tt.sortKeys}
			result := CompareWithOptions(left, right, opts)

			if result.Root.Type != tt.expectedType {
				t.Errorf("expected root type %q, got %q", tt.expectedType, result.Root.Type)
			}

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v (type=%s)",
					tt.expectEqual, isEqual, result.Root.Type)
			}
		})
	}
}

// TestCompareWithOptionsNormalizeNumbers tests the Numbers checkbox
func TestCompareWithOptionsNormalizeNumbers(t *testing.T) {
	tests := []struct {
		name             string
		leftJSON         string
		rightJSON        string
		normalizeNumbers bool
		expectEqual      bool
		expectedType     DiffType
	}{
		{
			name:             "1.0 vs 1 - normalize enabled - should be equal",
			leftJSON:        `1.0`,
			rightJSON:       `1`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
		{
			name:             "nested number normalization",
			leftJSON:        `{"value": 42.0}`,
			rightJSON:       `{"value": 42}`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
		{
			name:             "array of numbers",
			leftJSON:        `[1.0, 2.0, 3.0]`,
			rightJSON:       `[1, 2, 3]`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
		{
			name:             "decimal vs integer - not equal",
			leftJSON:        `1.5`,
			rightJSON:       `1`,
			normalizeNumbers: true,
			expectEqual:      false,
			expectedType:     DiffChanged,
		},
		{
			name:             "zero normalization",
			leftJSON:        `0.0`,
			rightJSON:       `0`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
		{
			name:             "negative whole numbers",
			leftJSON:        `-5.0`,
			rightJSON:       `-5`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
		{
			name:             "complex nested structure",
			leftJSON:        `{"users": [{"id": 1.0, "score": 95.0}]}`,
			rightJSON:       `{"users": [{"id": 1, "score": 95}]}`,
			normalizeNumbers: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			opts := normalize.Options{NormalizeNumbers: tt.normalizeNumbers}
			result := CompareWithOptions(left, right, opts)

			if result.Root.Type != tt.expectedType {
				t.Errorf("expected root type %q, got %q", tt.expectedType, result.Root.Type)
			}

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v", tt.expectEqual, isEqual)
			}
		})
	}
}

// TestCompareWithOptionsTrimStrings tests the Trim Strings checkbox
func TestCompareWithOptionsTrimStrings(t *testing.T) {
	tests := []struct {
		name         string
		leftJSON     string
		rightJSON    string
		trimStrings  bool
		expectEqual  bool
		expectedType DiffType
	}{
		{
			name:         "whitespace difference - trim enabled - should be equal",
			leftJSON:    `"  hello  "`,
			rightJSON:   `"hello"`,
			trimStrings:  true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "whitespace difference - trim disabled - should differ",
			leftJSON:    `"  hello  "`,
			rightJSON:   `"hello"`,
			trimStrings:  false,
			expectEqual:  false,
			expectedType: DiffChanged,
		},
		{
			name:         "nested string trimming",
			leftJSON:    `{"name": "  Alice  "}`,
			rightJSON:   `{"name": "Alice"}`,
			trimStrings:  true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "array of strings",
			leftJSON:    `["  a  ", "  b  "]`,
			rightJSON:   `["a", "b"]`,
			trimStrings:  true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "internal spaces preserved",
			leftJSON:    `"  hello   world  "`,
			rightJSON:   `"hello   world"`,
			trimStrings:  true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
		{
			name:         "different content - not equal",
			leftJSON:    `"  hello  "`,
			rightJSON:   `"  world  "`,
			trimStrings:  true,
			expectEqual:  false,
			expectedType: DiffChanged,
		},
		{
			name:         "tabs and newlines trimmed",
			leftJSON:    `"\thello\n"`,
			rightJSON:   `"hello"`,
			trimStrings:  true,
			expectEqual:  true,
			expectedType: DiffEqual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			opts := normalize.Options{TrimStrings: tt.trimStrings}
			result := CompareWithOptions(left, right, opts)

			if result.Root.Type != tt.expectedType {
				t.Errorf("expected root type %q, got %q", tt.expectedType, result.Root.Type)
			}

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v", tt.expectEqual, isEqual)
			}
		})
	}
}

// TestCompareWithOptionsNullEqualsAbsent tests the Null = Absent checkbox
func TestCompareWithOptionsNullEqualsAbsent(t *testing.T) {
	tests := []struct {
		name             string
		leftJSON         string
		rightJSON        string
		nullEqualsAbsent bool
		expectEqual      bool
		expectedType     DiffType
		expectedStats    DiffStats
	}{
		{
			name:             "null vs missing key - enabled - should be equal",
			leftJSON:        `{"a": 1, "b": null}`,
			rightJSON:       `{"a": 1}`,
			nullEqualsAbsent: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
			expectedStats:    DiffStats{Equal: 1},
		},
		{
			name:             "null vs missing key - disabled - should differ",
			leftJSON:        `{"a": 1, "b": null}`,
			rightJSON:       `{"a": 1}`,
			nullEqualsAbsent: false,
			expectEqual:      false,
			expectedType:     DiffChanged,
			expectedStats:    DiffStats{Equal: 1, Removed: 1},
		},
		{
			name:             "multiple nulls vs missing",
			leftJSON:        `{"a": 1, "b": null, "c": null}`,
			rightJSON:       `{"a": 1}`,
			nullEqualsAbsent: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
			expectedStats:    DiffStats{Equal: 1},
		},
		{
			name:             "nested null vs missing",
			leftJSON:        `{"outer": {"a": 1, "b": null}}`,
			rightJSON:       `{"outer": {"a": 1}}`,
			nullEqualsAbsent: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
			expectedStats:    DiffStats{Equal: 1},
		},
		{
			name:             "null vs actual value - should differ",
			leftJSON:        `{"a": null}`,
			rightJSON:       `{"a": 1}`,
			nullEqualsAbsent: true,
			expectEqual:      false,
			expectedType:     DiffChanged,
			expectedStats:    DiffStats{Added: 1},
		},
		{
			name:             "both have nulls - equal when disabled too",
			leftJSON:        `{"a": null}`,
			rightJSON:       `{"a": null}`,
			nullEqualsAbsent: false,
			expectEqual:      true,
			expectedType:     DiffEqual,
			expectedStats:    DiffStats{Equal: 1},
		},
		{
			name:             "array with null objects",
			leftJSON:        `[{"id": 1, "extra": null}]`,
			rightJSON:       `[{"id": 1}]`,
			nullEqualsAbsent: true,
			expectEqual:      true,
			expectedType:     DiffEqual,
			expectedStats:    DiffStats{Equal: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			opts := normalize.Options{NullEqualsAbsent: tt.nullEqualsAbsent}
			result := CompareWithOptions(left, right, opts)

			if result.Root.Type != tt.expectedType {
				t.Errorf("expected root type %q, got %q", tt.expectedType, result.Root.Type)
			}

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v", tt.expectEqual, isEqual)
			}

			if result.Stats != tt.expectedStats {
				t.Errorf("expected stats %+v, got %+v", tt.expectedStats, result.Stats)
			}
		})
	}
}

// TestCompareWithOptionsCombined tests all options together
func TestCompareWithOptionsCombined(t *testing.T) {
	tests := []struct {
		name          string
		leftJSON      string
		rightJSON     string
		opts          normalize.Options
		expectEqual   bool
		expectedStats DiffStats
	}{
		{
			name:     "all normalizations combined - complex",
			leftJSON: `{"z": 1.0, "a": "  hello  ", "n": null, "b": 2.0}`,
			rightJSON: `{"a": "hello", "b": 2, "z": 1}`,
			opts: normalize.Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			expectEqual:   true,
			expectedStats: DiffStats{Equal: 3},
		},
		{
			name:     "nested with all options",
			leftJSON: `{"outer": {"z": 100.0, "a": "  value  ", "x": null}}`,
			rightJSON: `{"outer": {"a": "value", "z": 100}}`,
			opts: normalize.Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			expectEqual:   true,
			expectedStats: DiffStats{Equal: 2},
		},
		{
			name:     "some options disabled - should differ",
			leftJSON: `{"a": 1.0, "b": null}`,
			rightJSON: `{"a": 1}`,
			opts: normalize.Options{
				NormalizeNumbers: true,
				NullEqualsAbsent: false, // null treated differently
			},
			expectEqual:   false,
			expectedStats: DiffStats{Equal: 1, Removed: 1},
		},
		{
			name:     "real-world API response comparison",
			leftJSON: `{"user": {"id": 123.0, "name": "  John Doe  ", "deleted_at": null}, "status": "active"}`,
			rightJSON: `{"status": "active", "user": {"name": "John Doe", "id": 123}}`,
			opts: normalize.Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			expectEqual:   true,
			expectedStats: DiffStats{Equal: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			result := CompareWithOptions(left, right, tt.opts)

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v (type=%s)",
					tt.expectEqual, isEqual, result.Root.Type)
			}

			if result.Stats != tt.expectedStats {
				t.Errorf("expected stats %+v, got %+v", tt.expectedStats, result.Stats)
			}
		})
	}
}

// TestCompareWithOptionsEdgeCases tests edge cases
func TestCompareWithOptionsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		leftJSON    string
		rightJSON   string
		opts        normalize.Options
		expectEqual bool
	}{
		{
			name:        "empty objects",
			leftJSON:   `{}`,
			rightJSON:  `{}`,
			opts:        normalize.DefaultOptions(),
			expectEqual: true,
		},
		{
			name:        "empty arrays",
			leftJSON:   `[]`,
			rightJSON:  `[]`,
			opts:        normalize.DefaultOptions(),
			expectEqual: true,
		},
		{
			name:        "deeply nested - 5 levels",
			leftJSON:   `{"a":{"b":{"c":{"d":{"e":1.0}}}}}`,
			rightJSON:  `{"a":{"b":{"c":{"d":{"e":1}}}}}`,
			opts:        normalize.Options{NormalizeNumbers: true},
			expectEqual: true,
		},
		{
			name:        "unicode strings with whitespace",
			leftJSON:   `{"name": "  日本語  "}`,
			rightJSON:  `{"name": "日本語"}`,
			opts:        normalize.Options{TrimStrings: true},
			expectEqual: true,
		},
		{
			name:        "mixed array types",
			leftJSON:   `[1.0, "  text  ", null, true]`,
			rightJSON:  `[1, "text", null, true]`,
			opts:        normalize.Options{NormalizeNumbers: true, TrimStrings: true},
			expectEqual: true,
		},
		{
			name:        "object with all null values",
			leftJSON:   `{"a": null, "b": null, "c": null}`,
			rightJSON:  `{}`,
			opts:        normalize.Options{NullEqualsAbsent: true},
			expectEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.leftJSON), &left)
			json.Unmarshal([]byte(tt.rightJSON), &right)

			result := CompareWithOptions(left, right, tt.opts)

			isEqual := result.Root.Type == DiffEqual
			if isEqual != tt.expectEqual {
				t.Errorf("expected equal=%v, got equal=%v", tt.expectEqual, isEqual)
			}
		})
	}
}
