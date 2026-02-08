// Package normalize provides JSON normalization for semantic comparison.
//
// Normalization transforms JSON values so that semantically equivalent
// values become structurally identical, making diff results more meaningful.
package normalize

// Options controls how JSON normalization is performed.
type Options struct {
	// SortKeys sorts object keys alphabetically.
	// When true: {"b":1, "a":2} is normalized to {"a":2, "b":1}
	// This is the most common normalization - key order rarely matters in JSON.
	SortKeys bool

	// NormalizeNumbers converts all numbers to a canonical form.
	// When true: 1.0 becomes 1, 1.00000 becomes 1
	// Handles floating point representation differences.
	NormalizeNumbers bool

	// TrimStrings removes leading/trailing whitespace from strings.
	// When true: "  hello  " becomes "hello"
	// Use with caution - whitespace may be significant in some contexts.
	TrimStrings bool

	// NullEqualsAbsent treats null values as equivalent to missing keys.
	// When true: {"a": null} is equivalent to {}
	// Useful for APIs that inconsistently include/omit null fields.
	NullEqualsAbsent bool

	// SortArrays sorts array elements.
	// When true: [3, 1, 2] becomes [1, 2, 3]
	// WARNING: Only use if array order truly doesn't matter in your data!
	// For arrays of objects, use SortArraysByKey instead.
	SortArrays bool

	// SortArraysByKey sorts arrays of objects by a specific key.
	// Example: With SortArraysByKey="id",
	//   [{"id":2}, {"id":1}] becomes [{"id":1}, {"id":2}]
	// Empty string means don't sort by key.
	SortArraysByKey string
}

// DefaultOptions returns sensible defaults for normalization.
func DefaultOptions() Options {
	return Options{
		SortKeys:         true,  // Almost always wanted
		NormalizeNumbers: true,  // Safe default
		TrimStrings:      false, // Could change semantics
		NullEqualsAbsent: false, // Could hide real differences
		SortArrays:       false, // Order usually matters
		SortArraysByKey:  "",    // Disabled by default
	}
}

// NoNormalization returns options with all normalization disabled.
// Useful for exact/strict comparison.
func NoNormalization() Options {
	return Options{
		SortKeys:         false,
		NormalizeNumbers: false,
		TrimStrings:      false,
		NullEqualsAbsent: false,
		SortArrays:       false,
		SortArraysByKey:  "",
	}
}
