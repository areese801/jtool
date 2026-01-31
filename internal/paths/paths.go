// Package paths extracts all JSON paths from a JSON structure.
// This is useful for understanding the schema/structure of a JSON document,
// especially when dealing with arrays of objects where paths repeat.
//
// Python comparison:
//   - Similar to recursively walking a dict/list structure
//   - Go requires explicit type assertions; Python uses duck typing
//   - Go's map iteration order is random; we sort at the end
package paths

import (
	"sort"
)

// PathInfo holds information about a JSON path.
type PathInfo struct {
	Path  string `json:"path"`  // The JSON path (e.g., "$.users.name")
	Count int    `json:"count"` // How many times this path appears (for arrays)
}

// PathResult is the result of extracting paths from JSON.
type PathResult struct {
	Paths      []PathInfo `json:"paths"`      // All paths found, sorted alphabetically
	TotalPaths int        `json:"totalPaths"` // Total number of unique paths
	TotalLeafs int        `json:"totalLeafs"` // Total leaf values (sum of counts)
}

// Extract walks a JSON structure and returns all paths to leaf values.
//
// Python equivalent:
//
//	def extract(data):
//	    paths = defaultdict(int)
//	    unpack("$", data, paths)
//	    return sorted(paths.items())
//
// Key Go difference: We use type switches instead of isinstance() checks.
func Extract(data any) *PathResult {
	// Map to count occurrences of each path
	// Python: paths = defaultdict(int)
	pathCounts := make(map[string]int)

	// Recursively extract paths (starting with empty prefix for jq compatibility)
	extractPaths("", data, pathCounts)

	// Convert map to sorted slice
	// Go maps have random iteration order, so we must sort explicitly
	// Python's OrderedDict or sorted() handles this
	paths := make([]PathInfo, 0, len(pathCounts))
	totalLeafs := 0

	for path, count := range pathCounts {
		paths = append(paths, PathInfo{Path: path, Count: count})
		totalLeafs += count
	}

	// Sort alphabetically by path
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})

	return &PathResult{
		Paths:      paths,
		TotalPaths: len(paths),
		TotalLeafs: totalLeafs,
	}
}

// extractPaths recursively walks the JSON structure.
//
// Python equivalent:
//
//	def unpack(pre_path, test_obj, result_set):
//	    if isinstance(test_obj, dict):
//	        for k, v in test_obj.items():
//	            # ... recurse
//	    elif isinstance(test_obj, list):
//	        for item in test_obj:
//	            # ... recurse
//
// Go uses type switch instead of isinstance():
//
//	switch v := value.(type) {
//	case map[string]any:  // like isinstance(v, dict)
//	case []any:           // like isinstance(v, list)
//	}
func extractPaths(prefix string, value any, counts map[string]int) {
	switch v := value.(type) {
	case map[string]any:
		// Object: recurse into each key
		// Python: if isinstance(test_obj, dict):
		for key, val := range v {
			childPath := prefix + "." + key
			extractPaths(childPath, val, counts)
		}

	case []any:
		// Array: use [] notation to indicate array element
		// Python: if isinstance(test_obj, list):
		// Original script used "." for arrays; we'll use "[]" for clarity
		for _, item := range v {
			childPath := prefix + "[]"
			extractPaths(childPath, item, counts)
		}

	default:
		// Leaf value (string, number, bool, null)
		// This is an atomic value - record the path
		// Python: result_set[json_path] += 1
		counts[prefix]++
	}
}
