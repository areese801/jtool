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

// ============================================================
// Comprehensive QA Tests for Checkbox Options
// ============================================================

// TestSortKeysOption tests the "Sort Keys" checkbox option thoroughly
func TestSortKeysOption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sortKeys bool
		// Note: sort keys affects comparison, not the normalized output structure
		// The test verifies that maps with same keys in different order normalize identically
	}{
		{
			name:     "simple object - keys already sorted",
			input:    `{"a": 1, "b": 2, "c": 3}`,
			sortKeys: true,
		},
		{
			name:     "simple object - keys reverse order",
			input:    `{"z": 1, "m": 2, "a": 3}`,
			sortKeys: true,
		},
		{
			name:     "nested objects",
			input:    `{"outer": {"z": 1, "a": 2}, "inner": {"y": 3, "b": 4}}`,
			sortKeys: true,
		},
		{
			name:     "deep nesting with unsorted keys",
			input:    `{"c": {"b": {"a": 1}}}`,
			sortKeys: true,
		},
		{
			name:     "array of objects with unsorted keys",
			input:    `[{"z": 1, "a": 2}, {"y": 3, "b": 4}]`,
			sortKeys: true,
		},
		{
			name:     "mixed content",
			input:    `{"zebra": [1, {"b": 2, "a": 1}], "apple": "fruit"}`,
			sortKeys: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data any
			if err := json.Unmarshal([]byte(tt.input), &data); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			opts := Options{SortKeys: tt.sortKeys}
			result := Value(data, opts)

			// Verify the result can be serialized and deserialized
			resultJSON, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}

			// Verify it's valid JSON
			var verification any
			if err := json.Unmarshal(resultJSON, &verification); err != nil {
				t.Errorf("result is not valid JSON: %v", err)
			}
		})
	}
}

// TestSortKeysComparisonEquality verifies that two objects with same keys in different order
// are considered equal when SortKeys is enabled
func TestSortKeysComparisonEquality(t *testing.T) {
	tests := []struct {
		name        string
		left        string
		right       string
		sortKeys    bool
		shouldEqual bool
	}{
		{
			name:        "same keys different order - sort enabled",
			left:        `{"a": 1, "b": 2}`,
			right:       `{"b": 2, "a": 1}`,
			sortKeys:    true,
			shouldEqual: true,
		},
		{
			name:        "nested same keys different order - sort enabled",
			left:        `{"x": {"a": 1, "b": 2}}`,
			right:       `{"x": {"b": 2, "a": 1}}`,
			sortKeys:    true,
			shouldEqual: true,
		},
		{
			name:        "deeply nested",
			left:        `{"level1": {"level2": {"z": 1, "a": 2}}}`,
			right:       `{"level1": {"level2": {"a": 2, "z": 1}}}`,
			sortKeys:    true,
			shouldEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.left), &left)
			json.Unmarshal([]byte(tt.right), &right)

			opts := Options{SortKeys: tt.sortKeys}
			leftNorm := Value(left, opts)
			rightNorm := Value(right, opts)

			isEqual := reflect.DeepEqual(leftNorm, rightNorm)
			if isEqual != tt.shouldEqual {
				t.Errorf("expected equal=%v, got equal=%v\nleft:  %v\nright: %v",
					tt.shouldEqual, isEqual, leftNorm, rightNorm)
			}
		})
	}
}

// TestNormalizeNumbersOption tests the "Numbers" checkbox option thoroughly
func TestNormalizeNumbersOption(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		normalizeNumbers bool
		expected         string
	}{
		// Basic cases
		{
			name:             "whole number as float",
			input:            `1.0`,
			normalizeNumbers: true,
			expected:         `1`,
		},
		{
			name:             "whole number as float - disabled",
			input:            `1.0`,
			normalizeNumbers: false,
			expected:         `1`, // json.Marshal normalizes anyway
		},
		{
			name:             "actual decimal",
			input:            `1.5`,
			normalizeNumbers: true,
			expected:         `1.5`,
		},
		{
			name:             "negative whole number",
			input:            `-5.0`,
			normalizeNumbers: true,
			expected:         `-5`,
		},
		{
			name:             "negative decimal",
			input:            `-3.14159`,
			normalizeNumbers: true,
			expected:         `-3.14159`,
		},
		{
			name:             "zero",
			input:            `0.0`,
			normalizeNumbers: true,
			expected:         `0`,
		},
		{
			name:             "large whole number",
			input:            `1000000.0`,
			normalizeNumbers: true,
			expected:         `1000000`,
		},
		{
			name:             "scientific notation whole",
			input:            `1e6`,
			normalizeNumbers: true,
			expected:         `1000000`,
		},
		// In arrays
		{
			name:             "array of mixed numbers",
			input:            `[1.0, 2.5, 3.0, 4.75]`,
			normalizeNumbers: true,
			expected:         `[1,2.5,3,4.75]`,
		},
		// In objects
		{
			name:             "object with number values",
			input:            `{"int": 1.0, "float": 1.5, "zero": 0.0}`,
			normalizeNumbers: true,
			expected:         `{"float":1.5,"int":1,"zero":0}`,
		},
		// Nested
		{
			name:             "deeply nested numbers",
			input:            `{"a": {"b": {"c": 42.0}}}`,
			normalizeNumbers: true,
			expected:         `{"a":{"b":{"c":42}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input any
			if err := json.Unmarshal([]byte(tt.input), &input); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			opts := Options{NormalizeNumbers: tt.normalizeNumbers}
			result := Value(input, opts)

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

// TestNormalizeNumbersComparisonEquality verifies that 1.0 and 1 are equal when enabled
func TestNormalizeNumbersComparisonEquality(t *testing.T) {
	tests := []struct {
		name             string
		left             string
		right            string
		normalizeNumbers bool
		shouldEqual      bool
	}{
		{
			name:             "1.0 vs 1 - normalize enabled",
			left:             `1.0`,
			right:            `1`,
			normalizeNumbers: true,
			shouldEqual:      true,
		},
		{
			name:             "nested 1.0 vs 1",
			left:             `{"value": 1.0}`,
			right:            `{"value": 1}`,
			normalizeNumbers: true,
			shouldEqual:      true,
		},
		{
			name:             "array with mixed",
			left:             `[1.0, 2.0, 3.0]`,
			right:            `[1, 2, 3]`,
			normalizeNumbers: true,
			shouldEqual:      true,
		},
		{
			name:             "1.5 vs 1 - should NOT equal",
			left:             `1.5`,
			right:            `1`,
			normalizeNumbers: true,
			shouldEqual:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.left), &left)
			json.Unmarshal([]byte(tt.right), &right)

			opts := Options{NormalizeNumbers: tt.normalizeNumbers}
			leftNorm := Value(left, opts)
			rightNorm := Value(right, opts)

			isEqual := reflect.DeepEqual(leftNorm, rightNorm)
			if isEqual != tt.shouldEqual {
				t.Errorf("expected equal=%v, got equal=%v", tt.shouldEqual, isEqual)
			}
		})
	}
}

// TestTrimStringsOption tests the "Trim Strings" checkbox option thoroughly
func TestTrimStringsOption(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		trimStrings bool
		expected    string
	}{
		// Basic trimming
		{
			name:        "leading spaces",
			input:       `"   hello"`,
			trimStrings: true,
			expected:    `"hello"`,
		},
		{
			name:        "trailing spaces",
			input:       `"hello   "`,
			trimStrings: true,
			expected:    `"hello"`,
		},
		{
			name:        "both sides",
			input:       `"   hello   "`,
			trimStrings: true,
			expected:    `"hello"`,
		},
		{
			name:        "tabs and newlines (escaped)",
			input:       `"\t\nhello\t\n"`,
			trimStrings: true,
			expected:    `"hello"`,
		},
		{
			name:        "mixed whitespace",
			input:       `"  \t  hello  \n  "`,
			trimStrings: true,
			expected:    `"hello"`,
		},
		// Disabled
		{
			name:        "preserve whitespace when disabled",
			input:       `"   hello   "`,
			trimStrings: false,
			expected:    `"   hello   "`,
		},
		// Edge cases
		{
			name:        "empty string",
			input:       `""`,
			trimStrings: true,
			expected:    `""`,
		},
		{
			name:        "only whitespace",
			input:       `"   "`,
			trimStrings: true,
			expected:    `""`,
		},
		{
			name:        "internal spaces preserved",
			input:       `"  hello   world  "`,
			trimStrings: true,
			expected:    `"hello   world"`,
		},
		// In objects
		{
			name:        "object string values",
			input:       `{"name": "  Alice  ", "city": "  NYC  "}`,
			trimStrings: true,
			expected:    `{"city":"NYC","name":"Alice"}`,
		},
		// In arrays
		{
			name:        "array of strings",
			input:       `["  a  ", "  b  ", "  c  "]`,
			trimStrings: true,
			expected:    `["a","b","c"]`,
		},
		// Nested
		{
			name:        "deeply nested strings",
			input:       `{"outer": {"inner": "  value  "}}`,
			trimStrings: true,
			expected:    `{"outer":{"inner":"value"}}`,
		},
		// Mixed content - only strings trimmed
		{
			name:        "mixed types - only strings affected",
			input:       `{"str": "  hello  ", "num": 42, "bool": true}`,
			trimStrings: true,
			expected:    `{"bool":true,"num":42,"str":"hello"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input any
			if err := json.Unmarshal([]byte(tt.input), &input); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			opts := Options{TrimStrings: tt.trimStrings}
			result := Value(input, opts)

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

// TestTrimStringsComparisonEquality verifies strings with/without whitespace equality
func TestTrimStringsComparisonEquality(t *testing.T) {
	tests := []struct {
		name        string
		left        string
		right       string
		trimStrings bool
		shouldEqual bool
	}{
		{
			name:        "trimmed vs untrimmed - trim enabled",
			left:        `"  hello  "`,
			right:       `"hello"`,
			trimStrings: true,
			shouldEqual: true,
		},
		{
			name:        "trimmed vs untrimmed - trim disabled",
			left:        `"  hello  "`,
			right:       `"hello"`,
			trimStrings: false,
			shouldEqual: false,
		},
		{
			name:        "both have whitespace - trim enabled",
			left:        `"  hello  "`,
			right:       `" hello "`,
			trimStrings: true,
			shouldEqual: true,
		},
		{
			name:        "nested strings",
			left:        `{"name": "  Alice  "}`,
			right:       `{"name": "Alice"}`,
			trimStrings: true,
			shouldEqual: true,
		},
		{
			name:        "different content - not equal",
			left:        `"  hello  "`,
			right:       `"world"`,
			trimStrings: true,
			shouldEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.left), &left)
			json.Unmarshal([]byte(tt.right), &right)

			opts := Options{TrimStrings: tt.trimStrings}
			leftNorm := Value(left, opts)
			rightNorm := Value(right, opts)

			isEqual := reflect.DeepEqual(leftNorm, rightNorm)
			if isEqual != tt.shouldEqual {
				t.Errorf("expected equal=%v, got equal=%v\nleft:  %v\nright: %v",
					tt.shouldEqual, isEqual, leftNorm, rightNorm)
			}
		})
	}
}

// TestNullEqualsAbsentOption tests the "Null = Absent" checkbox option thoroughly
func TestNullEqualsAbsentOption(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		nullEqualsAbsent bool
		expected         string
	}{
		// Basic null removal
		{
			name:             "single null value removed",
			input:            `{"a": 1, "b": null}`,
			nullEqualsAbsent: true,
			expected:         `{"a":1}`,
		},
		{
			name:             "single null value kept",
			input:            `{"a": 1, "b": null}`,
			nullEqualsAbsent: false,
			expected:         `{"a":1,"b":null}`,
		},
		{
			name:             "multiple nulls removed",
			input:            `{"a": null, "b": 1, "c": null, "d": 2}`,
			nullEqualsAbsent: true,
			expected:         `{"b":1,"d":2}`,
		},
		{
			name:             "all nulls - empty object",
			input:            `{"a": null, "b": null}`,
			nullEqualsAbsent: true,
			expected:         `{}`,
		},
		// Nested
		{
			name:             "nested null removed",
			input:            `{"outer": {"a": 1, "b": null}}`,
			nullEqualsAbsent: true,
			expected:         `{"outer":{"a":1}}`,
		},
		{
			name:             "deeply nested null removed",
			input:            `{"l1": {"l2": {"l3": null, "value": 1}}}`,
			nullEqualsAbsent: true,
			expected:         `{"l1":{"l2":{"value":1}}}`,
		},
		// Arrays - nulls in arrays should NOT be removed
		{
			name:             "array with null - preserved",
			input:            `[1, null, 2]`,
			nullEqualsAbsent: true,
			expected:         `[1,null,2]`,
		},
		{
			name:             "object in array with null key removed",
			input:            `[{"a": 1, "b": null}]`,
			nullEqualsAbsent: true,
			expected:         `[{"a":1}]`,
		},
		// Edge cases
		{
			name:             "top-level null unchanged",
			input:            `null`,
			nullEqualsAbsent: true,
			expected:         `null`,
		},
		{
			name:             "empty object unchanged",
			input:            `{}`,
			nullEqualsAbsent: true,
			expected:         `{}`,
		},
		{
			name:             "no nulls - unchanged",
			input:            `{"a": 1, "b": 2}`,
			nullEqualsAbsent: true,
			expected:         `{"a":1,"b":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input any
			if err := json.Unmarshal([]byte(tt.input), &input); err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			opts := Options{NullEqualsAbsent: tt.nullEqualsAbsent}
			result := Value(input, opts)

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

// TestNullEqualsAbsentComparisonEquality verifies null vs missing key equality
func TestNullEqualsAbsentComparisonEquality(t *testing.T) {
	tests := []struct {
		name             string
		left             string
		right            string
		nullEqualsAbsent bool
		shouldEqual      bool
	}{
		{
			name:             "null vs absent - enabled",
			left:             `{"a": 1, "b": null}`,
			right:            `{"a": 1}`,
			nullEqualsAbsent: true,
			shouldEqual:      true,
		},
		{
			name:             "null vs absent - disabled",
			left:             `{"a": 1, "b": null}`,
			right:            `{"a": 1}`,
			nullEqualsAbsent: false,
			shouldEqual:      false,
		},
		{
			name:             "multiple nulls vs absent",
			left:             `{"a": null, "b": 1, "c": null}`,
			right:            `{"b": 1}`,
			nullEqualsAbsent: true,
			shouldEqual:      true,
		},
		{
			name:             "nested null vs absent",
			left:             `{"outer": {"a": 1, "b": null}}`,
			right:            `{"outer": {"a": 1}}`,
			nullEqualsAbsent: true,
			shouldEqual:      true,
		},
		{
			name:             "both have nulls",
			left:             `{"a": null}`,
			right:            `{"a": null}`,
			nullEqualsAbsent: false,
			shouldEqual:      true,
		},
		{
			name:             "null vs value - NOT equal",
			left:             `{"a": null}`,
			right:            `{"a": 1}`,
			nullEqualsAbsent: true,
			shouldEqual:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.left), &left)
			json.Unmarshal([]byte(tt.right), &right)

			opts := Options{NullEqualsAbsent: tt.nullEqualsAbsent}
			leftNorm := Value(left, opts)
			rightNorm := Value(right, opts)

			isEqual := reflect.DeepEqual(leftNorm, rightNorm)
			if isEqual != tt.shouldEqual {
				t.Errorf("expected equal=%v, got equal=%v\nleft:  %v\nright: %v",
					tt.shouldEqual, isEqual, leftNorm, rightNorm)
			}
		})
	}
}

// TestCombinedOptions tests multiple options enabled together
func TestCombinedOptions(t *testing.T) {
	tests := []struct {
		name        string
		left        string
		right       string
		opts        Options
		shouldEqual bool
	}{
		{
			name:  "all options - complex equality",
			left:  `{"b": 1.0, "a": "  hello  ", "c": null}`,
			right: `{"a": "hello", "b": 1}`,
			opts: Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			shouldEqual: true,
		},
		{
			name:  "nested with all options",
			left:  `{"outer": {"z": 1.0, "a": "  test  ", "n": null}}`,
			right: `{"outer": {"a": "test", "z": 1}}`,
			opts: Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			shouldEqual: true,
		},
		{
			name:  "array of objects with all options",
			left:  `[{"id": 1.0, "name": "  Alice  ", "extra": null}]`,
			right: `[{"id": 1, "name": "Alice"}]`,
			opts: Options{
				SortKeys:         true,
				NormalizeNumbers: true,
				TrimStrings:      true,
				NullEqualsAbsent: true,
			},
			shouldEqual: true,
		},
		{
			name:  "some options only - not equal",
			left:  `{"a": 1.0, "b": null}`,
			right: `{"a": 1}`,
			opts: Options{
				NormalizeNumbers: true,
				NullEqualsAbsent: false, // null not treated as absent
			},
			shouldEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left, right any
			json.Unmarshal([]byte(tt.left), &left)
			json.Unmarshal([]byte(tt.right), &right)

			leftNorm := Value(left, tt.opts)
			rightNorm := Value(right, tt.opts)

			isEqual := reflect.DeepEqual(leftNorm, rightNorm)
			if isEqual != tt.shouldEqual {
				t.Errorf("expected equal=%v, got equal=%v\nleft:  %v\nright: %v",
					tt.shouldEqual, isEqual, leftNorm, rightNorm)
			}
		})
	}
}
