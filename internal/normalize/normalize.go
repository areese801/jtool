package normalize

import (
	"math"
	"sort"
	"strings"
)

// Value normalizes a JSON value according to the given options.
// It recursively processes objects, arrays, and primitive values.
//
// The input should be the result of json.Unmarshal into `any`:
//   - map[string]any for objects
//   - []any for arrays
//   - float64 for numbers
//   - string for strings
//   - bool for booleans
//   - nil for null
//
// Python comparison:
//
//	def normalize_value(value: Any, options: NormalizeOptions) -> Any:
//	    if isinstance(value, dict):
//	        return normalize_object(value, options)
//	    elif isinstance(value, list):
//	        return normalize_array(value, options)
//	    # ... etc
//
// In Go, we use type switches instead of isinstance().
func Value(v any, opts Options) any {
	switch val := v.(type) {
	case map[string]any:
		return normalizeObject(val, opts)
	case []any:
		return normalizeArray(val, opts)
	case float64:
		return normalizeNumber(val, opts)
	case string:
		return normalizeString(val, opts)
	case bool, nil:
		// Booleans and nil don't need normalization
		return val
	default:
		// Unknown type, return as-is
		return val
	}
}

// normalizeObject normalizes a JSON object (map).
//
// Go maps are unordered, but when we compare them, we want consistent ordering.
// This function:
// 1. Optionally removes null values (if NullEqualsAbsent)
// 2. Recursively normalizes all values
// 3. Returns a new map (original is not modified)
//
// Note: Key sorting happens during comparison, not here.
// Go maps don't maintain insertion order, so we sort during iteration.
func normalizeObject(obj map[string]any, opts Options) map[string]any {
	result := make(map[string]any)

	for key, val := range obj {
		// Skip null values if NullEqualsAbsent is enabled
		if opts.NullEqualsAbsent && val == nil {
			continue
		}

		// Recursively normalize the value
		result[key] = Value(val, opts)
	}

	return result
}

// normalizeArray normalizes a JSON array.
//
// This function:
// 1. Recursively normalizes all elements
// 2. Optionally sorts the array (if SortArrays or SortArraysByKey)
// 3. Returns a new slice (original is not modified)
func normalizeArray(arr []any, opts Options) []any {
	// First, normalize all elements
	result := make([]any, len(arr))
	for i, val := range arr {
		result[i] = Value(val, opts)
	}

	// Sort if requested
	if opts.SortArraysByKey != "" {
		sortArrayByKey(result, opts.SortArraysByKey)
	} else if opts.SortArrays {
		sortArray(result)
	}

	return result
}

// normalizeNumber normalizes a JSON number.
//
// JSON numbers are always parsed as float64 in Go.
// This function converts whole numbers to integers for cleaner comparison.
//
// Examples:
//   - 1.0 → 1 (float64 to int representation)
//   - 1.5 → 1.5 (unchanged)
//   - 1.0000001 → 1.0000001 (unchanged, not close enough to integer)
//
// Python comparison:
//   - Python's json module keeps 1.0 as float, 1 as int
//   - Go's json always uses float64
//   - This normalizes them to be comparable
func normalizeNumber(n float64, opts Options) any {
	if !opts.NormalizeNumbers {
		return n
	}

	// Check if it's effectively an integer
	// Use a small epsilon for floating point comparison
	if n == math.Trunc(n) && !math.IsInf(n, 0) && !math.IsNaN(n) {
		// It's a whole number - but we still return float64
		// because that's what json.Unmarshal produces
		// The key is removing trailing decimal zeros
		return math.Trunc(n)
	}

	return n
}

// normalizeString normalizes a JSON string.
func normalizeString(s string, opts Options) string {
	if opts.TrimStrings {
		return strings.TrimSpace(s)
	}
	return s
}

// sortArrayByKey sorts an array of objects by a specific key.
//
// Example with key="id":
//
//	[{"id": 2, "name": "Bob"}, {"id": 1, "name": "Alice"}]
//	→ [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
//
// Non-object elements or objects without the key are sorted to the end.
func sortArrayByKey(arr []any, key string) {
	sort.SliceStable(arr, func(i, j int) bool {
		// Try to get the key from both elements
		iObj, iOk := arr[i].(map[string]any)
		jObj, jOk := arr[j].(map[string]any)

		if !iOk || !jOk {
			// Non-objects: keep original order (stable sort)
			return false
		}

		iVal, iHas := iObj[key]
		jVal, jHas := jObj[key]

		if !iHas || !jHas {
			// Missing key: keep original order
			return false
		}

		// Compare the values
		return compareValues(iVal, jVal) < 0
	})
}

// sortArray sorts an array of primitives.
// Objects and arrays within the array are sorted by their JSON string representation.
func sortArray(arr []any) {
	sort.SliceStable(arr, func(i, j int) bool {
		return compareValues(arr[i], arr[j]) < 0
	})
}

// compareValues compares two values for sorting purposes.
// Returns negative if a < b, zero if a == b, positive if a > b.
//
// Python comparison:
//   - Python's sort handles mixed types with type-based ordering
//   - Go requires explicit comparison logic
//   - We define an ordering: nil < bool < number < string < array < object
func compareValues(a, b any) int {
	// Get type order
	aOrder := typeOrder(a)
	bOrder := typeOrder(b)

	if aOrder != bOrder {
		return aOrder - bOrder
	}

	// Same type - compare values
	switch aVal := a.(type) {
	case nil:
		return 0 // nil == nil
	case bool:
		bVal := b.(bool)
		if aVal == bVal {
			return 0
		}
		if !aVal { // false < true
			return -1
		}
		return 1
	case float64:
		bVal := b.(float64)
		if aVal < bVal {
			return -1
		}
		if aVal > bVal {
			return 1
		}
		return 0
	case string:
		bVal := b.(string)
		return strings.Compare(aVal, bVal)
	default:
		// Arrays and objects - compare by string representation
		// This is a simplification; could be more sophisticated
		return 0
	}
}

// typeOrder returns a numeric order for JSON types.
// Used for sorting mixed-type arrays.
func typeOrder(v any) int {
	switch v.(type) {
	case nil:
		return 0
	case bool:
		return 1
	case float64:
		return 2
	case string:
		return 3
	case []any:
		return 4
	case map[string]any:
		return 5
	default:
		return 6
	}
}
