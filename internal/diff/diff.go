// Package diff provides JSON comparison functionality.
//
// This package implements a recursive tree comparison algorithm for JSON data.
// It produces a structured diff that tracks the JSON path of each difference.
//
// Python comparison:
//   - Similar to the `deepdiff` library in Python
//   - Go requires explicit type checking (type assertions) whereas Python uses duck typing
//   - Go's `any` type is like Python's dynamic typing, but you must assert types to use them
package diff

import (
	"fmt"
	"reflect"
	"sort"

	"jsondiff/internal/normalize"
)

// Compare performs a diff between two parsed JSON values.
// Both left and right should be the result of json.Unmarshal into `any`.
//
// Returns a DiffResult containing the full diff tree and statistics.
//
// Python equivalent would be:
//
//	def compare(left: Any, right: Any) -> DiffResult:
//	    root = compare_values(left, right, "$")
//	    stats = calculate_stats(root)
//	    return DiffResult(root=root, stats=stats)
func Compare(left, right any) *DiffResult {
	root := compareValues(left, right, "")
	stats := calculateStats(root)

	return &DiffResult{
		Root:  root,
		Stats: stats,
	}
}

// CompareWithOptions performs a diff with normalization applied first.
// This allows semantically equivalent JSON to be treated as equal.
//
// Example:
//
//	opts := normalize.DefaultOptions()
//	result := diff.CompareWithOptions(left, right, opts)
//
// With default options:
//   - {"b":1, "a":2} equals {"a":2, "b":1} (key order ignored)
//   - 1.0 equals 1 (number normalization)
//
// Python comparison:
//
//	def compare_with_options(left, right, options):
//	    left_normalized = normalize(left, options)
//	    right_normalized = normalize(right, options)
//	    return compare(left_normalized, right_normalized)
func CompareWithOptions(left, right any, opts normalize.Options) *DiffResult {
	// Normalize both values before comparison
	leftNorm := normalize.Value(left, opts)
	rightNorm := normalize.Value(right, opts)

	// Now compare the normalized values
	root := compareValues(leftNorm, rightNorm, "")
	stats := calculateStats(root)

	return &DiffResult{
		Root:  root,
		Stats: stats,
	}
}

// compareValues recursively compares two values and returns a DiffNode.
// path is the JSON path to this value (e.g., "$.users[0].name").
//
// Go type assertions explained:
//   - In Python, you'd just access dict keys or list indices directly
//   - In Go, `any` (interface{}) could be anything, so we must assert the type
//   - leftMap, ok := left.(map[string]any) attempts to convert `left` to a map
//   - If successful, ok is true and leftMap contains the map
//   - If not, ok is false and leftMap is the zero value (nil for maps)
func compareValues(left, right any, path string) DiffNode {
	// Both nil/null - equal
	if left == nil && right == nil {
		return DiffNode{
			Path: path,
			Type: DiffEqual,
		}
	}

	// One is nil - added or removed
	if left == nil {
		return DiffNode{
			Path:  path,
			Type:  DiffAdded,
			Right: right,
		}
	}
	if right == nil {
		return DiffNode{
			Path: path,
			Type: DiffRemoved,
			Left: left,
		}
	}

	// Type assertion to check what kind of JSON value we have
	// json.Unmarshal produces these types:
	//   - map[string]any for objects
	//   - []any for arrays
	//   - float64 for numbers (always float64, even for integers!)
	//   - string for strings
	//   - bool for booleans
	//   - nil for null

	leftMap, leftIsMap := left.(map[string]any)
	rightMap, rightIsMap := right.(map[string]any)

	// Both are objects - compare recursively
	if leftIsMap && rightIsMap {
		return compareObjects(leftMap, rightMap, path)
	}

	leftArr, leftIsArr := left.([]any)
	rightArr, rightIsArr := right.([]any)

	// Both are arrays - compare element by element
	if leftIsArr && rightIsArr {
		return compareArrays(leftArr, rightArr, path)
	}

	// Different types - this is a change
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return DiffNode{
			Path:  path,
			Type:  DiffChanged,
			Left:  left,
			Right: right,
		}
	}

	// Same type, compare values directly
	// Using reflect.DeepEqual for safety (handles edge cases)
	//
	// Python comparison:
	//   - Python's == does deep equality for dicts/lists automatically
	//   - Go's == only works for comparable types (not slices/maps)
	//   - reflect.DeepEqual is Go's equivalent to Python's ==
	if reflect.DeepEqual(left, right) {
		return DiffNode{
			Path: path,
			Type: DiffEqual,
		}
	}

	return DiffNode{
		Path:  path,
		Type:  DiffChanged,
		Left:  left,
		Right: right,
	}
}

// compareObjects compares two JSON objects (maps) and returns a DiffNode.
//
// Algorithm:
//  1. Collect all keys from both maps
//  2. Sort keys for deterministic output
//  3. For each key:
//     - If only in left: removed
//     - If only in right: added
//     - If in both: recurse
func compareObjects(left, right map[string]any, path string) DiffNode {
	node := DiffNode{
		Path:     path,
		Type:     DiffEqual, // Will be updated if children have diffs
		Children: []DiffNode{},
	}

	// Collect all unique keys from both maps
	// Go maps don't have a union operation like Python's dict.keys() | other.keys()
	// We build a set manually using map[string]bool
	allKeys := make(map[string]bool)
	for k := range left {
		allKeys[k] = true
	}
	for k := range right {
		allKeys[k] = true
	}

	// Sort keys for deterministic output
	// In Python: sorted(all_keys)
	// In Go: we need to convert to slice first, then sort
	sortedKeys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Compare each key
	for _, key := range sortedKeys {
		childPath := fmt.Sprintf("%s.%s", path, key)

		leftVal, inLeft := left[key]
		rightVal, inRight := right[key]

		var child DiffNode
		if !inLeft {
			// Key only in right - added
			child = DiffNode{
				Path:  childPath,
				Type:  DiffAdded,
				Right: rightVal,
			}
		} else if !inRight {
			// Key only in left - removed
			child = DiffNode{
				Path: childPath,
				Type: DiffRemoved,
				Left: leftVal,
			}
		} else {
			// Key in both - recurse
			child = compareValues(leftVal, rightVal, childPath)
		}

		node.Children = append(node.Children, child)

		// Update parent type if any child is not equal
		if child.Type != DiffEqual {
			node.Type = DiffChanged
		}
	}

	return node
}

// compareArrays compares two JSON arrays element by element.
//
// Phase 1: Simple index-by-index comparison (order matters)
// Phase 2 (future): Could use LCS algorithm to detect moves
//
// Python comparison:
//   - In Python you might use zip_longest to iterate both lists
//   - In Go we manually track indices
func compareArrays(left, right []any, path string) DiffNode {
	node := DiffNode{
		Path:     path,
		Type:     DiffEqual,
		Children: []DiffNode{},
	}

	// Compare up to the length of the longer array
	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}

	for i := 0; i < maxLen; i++ {
		childPath := fmt.Sprintf("%s[%d]", path, i)

		var child DiffNode
		if i >= len(left) {
			// Index only in right - added
			child = DiffNode{
				Path:  childPath,
				Type:  DiffAdded,
				Right: right[i],
			}
		} else if i >= len(right) {
			// Index only in left - removed
			child = DiffNode{
				Path: childPath,
				Type: DiffRemoved,
				Left: left[i],
			}
		} else {
			// Index in both - recurse
			child = compareValues(left[i], right[i], childPath)
		}

		node.Children = append(node.Children, child)

		if child.Type != DiffEqual {
			node.Type = DiffChanged
		}
	}

	return node
}

// calculateStats walks the diff tree and counts each type of difference.
//
// Python equivalent:
//
//	def calculate_stats(node: DiffNode) -> DiffStats:
//	    stats = DiffStats()
//	    def walk(n):
//	        if n.type == "added": stats.added += 1
//	        # ... etc
//	        for child in n.children:
//	            walk(child)
//	    walk(node)
//	    return stats
func calculateStats(node DiffNode) DiffStats {
	stats := DiffStats{}
	walkAndCount(&stats, node)
	return stats
}

// walkAndCount recursively counts diff types.
// Note: we pass *DiffStats (pointer) so we modify the original, not a copy.
//
// Python comparison:
//   - Python passes objects by reference automatically
//   - Go passes by value, so we need a pointer to modify the original
//   - The & creates a pointer, the * dereferences it
func walkAndCount(stats *DiffStats, node DiffNode) {
	// Only count leaf nodes (nodes without children)
	// This avoids double-counting parent objects/arrays
	if len(node.Children) == 0 {
		switch node.Type {
		case DiffEqual:
			stats.Equal++
		case DiffAdded:
			stats.Added++
		case DiffRemoved:
			stats.Removed++
		case DiffChanged:
			stats.Changed++
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		walkAndCount(stats, child)
	}
}
