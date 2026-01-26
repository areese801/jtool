package normalize

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestNormalizeValue tests the main Value normalization function.
//
// Python comparison:
//
//	@pytest.mark.parametrize("input,options,expected", [...])
//	def test_normalize_value(input, options, expected):
//	    assert normalize(input, options) == expected
func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string // JSON input
		opts     Options
		expected string // Expected JSON output
	}{
		{
			name:     "number normalization - float to int",
			input:    `1.0`,
			opts:     Options{NormalizeNumbers: true},
			expected: `1`,
		},
		{
			name:     "number normalization - already int",
			input:    `42`,
			opts:     Options{NormalizeNumbers: true},
			expected: `42`,
		},
		{
			name:     "number normalization - decimal unchanged",
			input:    `1.5`,
			opts:     Options{NormalizeNumbers: true},
			expected: `1.5`,
		},
		{
			name:     "number normalization disabled",
			input:    `1.0`,
			opts:     Options{NormalizeNumbers: false},
			expected: `1`,  // Note: json.Marshal normalizes 1.0 to 1 anyway
		},
		{
			name:     "string trimming enabled",
			input:    `"  hello world  "`,
			opts:     Options{TrimStrings: true},
			expected: `"hello world"`,
		},
		{
			name:     "string trimming disabled",
			input:    `"  hello world  "`,
			opts:     Options{TrimStrings: false},
			expected: `"  hello world  "`,
		},
		{
			name:     "object unchanged without options",
			input:    `{"a": 1, "b": 2}`,
			opts:     NoNormalization(),
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "null equals absent - removes null",
			input:    `{"a": 1, "b": null}`,
			opts:     Options{NullEqualsAbsent: true},
			expected: `{"a":1}`,
		},
		{
			name:     "null equals absent disabled - keeps null",
			input:    `{"a": 1, "b": null}`,
			opts:     Options{NullEqualsAbsent: false},
			expected: `{"a":1,"b":null}`,
		},
		{
			name:     "nested object normalization",
			input:    `{"outer": {"inner": 1.0}}`,
			opts:     Options{NormalizeNumbers: true},
			expected: `{"outer":{"inner":1}}`,
		},
		{
			name:     "array normalization",
			input:    `[1.0, 2.0, 3.0]`,
			opts:     Options{NormalizeNumbers: true},
			expected: `[1,2,3]`,
		},
		{
			name:     "sort arrays - primitives",
			input:    `[3, 1, 2]`,
			opts:     Options{SortArrays: true},
			expected: `[1,2,3]`,
		},
		{
			name:     "sort arrays disabled",
			input:    `[3, 1, 2]`,
			opts:     Options{SortArrays: false},
			expected: `[3,1,2]`,
		},
		{
			name:     "boolean unchanged",
			input:    `true`,
			opts:     DefaultOptions(),
			expected: `true`,
		},
		{
			name:     "null unchanged",
			input:    `null`,
			opts:     DefaultOptions(),
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse input JSON
			var input any
			if err := json.Unmarshal([]byte(tt.input), &input); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			// Normalize
			result := Value(input, tt.opts)

			// Convert back to JSON for comparison
			resultJSON, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}

			if string(resultJSON) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(resultJSON))
			}
		})
	}
}

// TestSortArraysByKey tests sorting arrays of objects by a key
func TestSortArraysByKey(t *testing.T) {
	input := `[{"id": 3, "name": "C"}, {"id": 1, "name": "A"}, {"id": 2, "name": "B"}]`
	expected := `[{"id":1,"name":"A"},{"id":2,"name":"B"},{"id":3,"name":"C"}]`

	var data any
	json.Unmarshal([]byte(input), &data)

	opts := Options{SortArraysByKey: "id"}
	result := Value(data, opts)

	resultJSON, _ := json.Marshal(result)
	if string(resultJSON) != expected {
		t.Errorf("expected %s, got %s", expected, string(resultJSON))
	}
}

// TestDefaultOptions verifies default options are sensible
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.SortKeys {
		t.Error("expected SortKeys to be true by default")
	}
	if !opts.NormalizeNumbers {
		t.Error("expected NormalizeNumbers to be true by default")
	}
	if opts.TrimStrings {
		t.Error("expected TrimStrings to be false by default")
	}
	if opts.NullEqualsAbsent {
		t.Error("expected NullEqualsAbsent to be false by default")
	}
	if opts.SortArrays {
		t.Error("expected SortArrays to be false by default")
	}
}

// TestNoNormalization verifies no normalization options
func TestNoNormalization(t *testing.T) {
	opts := NoNormalization()

	if opts.SortKeys {
		t.Error("expected SortKeys to be false")
	}
	if opts.NormalizeNumbers {
		t.Error("expected NormalizeNumbers to be false")
	}
	if opts.TrimStrings {
		t.Error("expected TrimStrings to be false")
	}
}

// TestCompareValues tests the internal comparison function
func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		a, b     any
		expected int // -1, 0, or 1
	}{
		{"nil equals nil", nil, nil, 0},
		{"false < true", false, true, -1},
		{"true > false", true, false, 1},
		{"equal numbers", 1.0, 1.0, 0},
		{"less than number", 1.0, 2.0, -1},
		{"greater than number", 2.0, 1.0, 1},
		{"equal strings", "a", "a", 0},
		{"less than string", "a", "b", -1},
		{"greater than string", "b", "a", 1},
		{"different types", nil, 1.0, -1}, // nil < number
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)

			// Normalize to -1, 0, or 1
			var normalized int
			if result < 0 {
				normalized = -1
			} else if result > 0 {
				normalized = 1
			}

			if normalized != tt.expected {
				t.Errorf("compareValues(%v, %v) = %d, expected %d", tt.a, tt.b, normalized, tt.expected)
			}
		})
	}
}

// TestNormalizationIdempotent verifies normalizing twice gives same result
//
// This is a property-based test concept:
// normalize(normalize(x)) == normalize(x)
func TestNormalizationIdempotent(t *testing.T) {
	input := `{"b": 1.0, "a": [3, 1, 2], "c": "  test  "}`

	var data any
	json.Unmarshal([]byte(input), &data)

	opts := Options{
		NormalizeNumbers: true,
		TrimStrings:      true,
		SortArrays:       true,
	}

	// Normalize once
	first := Value(data, opts)

	// Normalize again
	second := Value(first, opts)

	// Should be identical
	if !reflect.DeepEqual(first, second) {
		t.Errorf("normalization is not idempotent:\nfirst:  %v\nsecond: %v", first, second)
	}
}
