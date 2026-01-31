package diff

import (
	"encoding/json"
	"testing"
)

// TestCompare uses table-driven tests, a common Go testing pattern.
//
// Python comparison:
//
//	# Python with pytest uses parametrize decorators
//	@pytest.mark.parametrize("left,right,expected_type", [
//	    ('{"a": 1}', '{"a": 1}', "equal"),
//	    ('{"a": 1}', '{"a": 2}', "changed"),
//	])
//	def test_compare(left, right, expected_type):
//	    result = compare(json.loads(left), json.loads(right))
//	    assert result.root.type == expected_type
//
// In Go, we use a slice of test cases and iterate over them.
// This gives us:
//   - Named test cases (shows in output)
//   - Easy to add new cases
//   - Can run individual cases with -run flag
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
