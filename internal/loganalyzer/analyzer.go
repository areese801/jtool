// Package loganalyzer parses files containing mixed content (logs + JSON)
// and aggregates JSON path statistics across all valid JSON lines.
//
// This is useful for analyzing Singer tap output, JSONL files, or any
// log file with embedded JSON objects.
//
// Python comparison:
//
//	for line in open(file):
//	    try:
//	        obj = json.loads(line)
//	        paths = get_json_paths(obj)
//	        aggregate(paths)
//	    except json.JSONDecodeError:
//	        pass  # Skip non-JSON lines
package loganalyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ValueFrequency represents a value and how often it appears.
type ValueFrequency struct {
	Value string `json:"value"` // The value (as string)
	Count int    `json:"count"` // How many times it appeared
}

// PathSummary holds aggregated information about a JSON path.
type PathSummary struct {
	Path          string           `json:"path"`          // The JSON path (e.g., "$.record.name")
	Count         int              `json:"count"`         // Total occurrences across all objects
	ObjectHits    int              `json:"objectHits"`    // Number of JSON objects containing this path
	DistinctCount int              `json:"distinctCount"` // Number of distinct values at this path
	TopValues     []ValueFrequency `json:"topValues"`     // Top 10 most frequent values
}

// AnalysisResult holds the complete analysis of a log file.
type AnalysisResult struct {
	Paths           []PathSummary `json:"paths"`           // All paths found, sorted by count desc
	TotalLines      int           `json:"totalLines"`      // Total lines in file
	JSONLines       int           `json:"jsonLines"`       // Lines that were valid JSON
	SkippedLines    int           `json:"skippedLines"`    // Lines that were not valid JSON
	TotalPaths      int           `json:"totalPaths"`      // Unique paths found
	TotalPathOccurs int           `json:"totalPathOccurs"` // Sum of all path counts
}

// AnalyzeFile reads a file and aggregates JSON path statistics.
//
// Python equivalent:
//
//	def analyze_file(path):
//	    paths = defaultdict(lambda: {"count": 0, "objects": 0})
//	    json_lines = 0
//	    for line in open(path):
//	        try:
//	            obj = json.loads(line.strip())
//	            json_lines += 1
//	            line_paths = extract_paths(obj)
//	            for p, count in line_paths.items():
//	                paths[p]["count"] += count
//	                paths[p]["objects"] += 1
//	        except:
//	            pass
//	    return paths, json_lines
func AnalyzeFile(filePath string) (*AnalysisResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Track path statistics
	// pathCounts[path] = total occurrences
	// pathObjects[path] = number of objects containing this path
	// pathValueFreq[path][value] = count of this value at this path
	pathCounts := make(map[string]int)
	pathObjects := make(map[string]int)
	pathValueFreq := make(map[string]map[string]int)

	totalLines := 0
	jsonLines := 0

	scanner := bufio.NewScanner(file)
	// Increase buffer size for long lines (default is 64KB, we'll use 1MB)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	// Multi-line JSON support: accumulate lines when we detect the start of
	// a JSON object/array that doesn't parse on a single line.
	// This handles pretty-printed JSON while keeping the fast path for JSONL.
	var accumulator strings.Builder
	inMultiLine := false
	const maxAccumulatorSize = 1024 * 1024 // 1MB safety limit

	// Helper to process a successfully parsed JSON object
	processJSON := func(data any) {
		jsonLines++
		linePathValues := make(map[string][]string)
		extractPathsWithValues("$", data, linePathValues)

		for path, values := range linePathValues {
			pathCounts[path] += len(values)
			pathObjects[path]++

			if pathValueFreq[path] == nil {
				pathValueFreq[path] = make(map[string]int)
			}
			for _, v := range values {
				pathValueFreq[path][v]++
			}
		}
	}

	for scanner.Scan() {
		totalLines++
		line := scanner.Text()

		if inMultiLine {
			// Continue accumulating lines
			accumulator.WriteString("\n")
			accumulator.WriteString(line)

			// Try to parse accumulated content
			var data any
			if err := json.Unmarshal([]byte(accumulator.String()), &data); err == nil {
				// Success! Process and reset
				processJSON(data)
				accumulator.Reset()
				inMultiLine = false
				continue
			}

			// Safety limit - abandon if too large
			if accumulator.Len() > maxAccumulatorSize {
				accumulator.Reset()
				inMultiLine = false
			}
			continue
		}

		// Fast path: try single-line parse first (works for JSONL)
		var data any
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			processJSON(data)
			continue
		}

		// Check if this might be the start of multi-line JSON
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			accumulator.WriteString(line)
			inMultiLine = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert to sorted slice (by count descending, then path ascending)
	paths := make([]PathSummary, 0, len(pathCounts))
	totalOccurs := 0

	for path, count := range pathCounts {
		valueFreqs := pathValueFreq[path]
		topValues := getTopValues(valueFreqs, 10)

		paths = append(paths, PathSummary{
			Path:          path,
			Count:         count,
			ObjectHits:    pathObjects[path],
			DistinctCount: len(valueFreqs),
			TopValues:     topValues,
		})
		totalOccurs += count
	}

	// Sort by count descending, then path ascending for ties
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Count != paths[j].Count {
			return paths[i].Count > paths[j].Count // Descending by count
		}
		return paths[i].Path < paths[j].Path // Ascending by path
	})

	return &AnalysisResult{
		Paths:           paths,
		TotalLines:      totalLines,
		JSONLines:       jsonLines,
		SkippedLines:    totalLines - jsonLines,
		TotalPaths:      len(paths),
		TotalPathOccurs: totalOccurs,
	}, nil
}

// AnalyzeString analyzes JSON lines from a string (for smaller inputs).
// Supports both JSONL (one object per line) and multi-line pretty-printed JSON.
func AnalyzeString(content string) (*AnalysisResult, error) {
	pathCounts := make(map[string]int)
	pathObjects := make(map[string]int)
	pathValueFreq := make(map[string]map[string]int)

	totalLines := 0
	jsonLines := 0

	// Multi-line JSON support
	var accumulator strings.Builder
	inMultiLine := false
	const maxAccumulatorSize = 1024 * 1024 // 1MB safety limit

	// Helper to process a successfully parsed JSON object
	processJSON := func(data any) {
		jsonLines++
		linePathValues := make(map[string][]string)
		extractPathsWithValues("$", data, linePathValues)

		for path, values := range linePathValues {
			pathCounts[path] += len(values)
			pathObjects[path]++

			if pathValueFreq[path] == nil {
				pathValueFreq[path] = make(map[string]int)
			}
			for _, v := range values {
				pathValueFreq[path][v]++
			}
		}
	}

	// Split by newlines and process each line
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		totalLines++

		if inMultiLine {
			// Continue accumulating lines
			accumulator.WriteString("\n")
			accumulator.WriteString(line)

			// Try to parse accumulated content
			var data any
			if err := json.Unmarshal([]byte(accumulator.String()), &data); err == nil {
				processJSON(data)
				accumulator.Reset()
				inMultiLine = false
				continue
			}

			// Safety limit
			if accumulator.Len() > maxAccumulatorSize {
				accumulator.Reset()
				inMultiLine = false
			}
			continue
		}

		// Fast path: try single-line parse first
		var data any
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			processJSON(data)
			continue
		}

		// Check if this might be the start of multi-line JSON
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			accumulator.WriteString(line)
			inMultiLine = true
		}
	}

	// Convert to sorted slice
	paths := make([]PathSummary, 0, len(pathCounts))
	totalOccurs := 0

	for path, count := range pathCounts {
		valueFreqs := pathValueFreq[path]
		topValues := getTopValues(valueFreqs, 10)

		paths = append(paths, PathSummary{
			Path:          path,
			Count:         count,
			ObjectHits:    pathObjects[path],
			DistinctCount: len(valueFreqs),
			TopValues:     topValues,
		})
		totalOccurs += count
	}

	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Count != paths[j].Count {
			return paths[i].Count > paths[j].Count
		}
		return paths[i].Path < paths[j].Path
	})

	return &AnalysisResult{
		Paths:           paths,
		TotalLines:      totalLines,
		JSONLines:       jsonLines,
		SkippedLines:    totalLines - jsonLines,
		TotalPaths:      len(paths),
		TotalPathOccurs: totalOccurs,
	}, nil
}

// extractPaths recursively extracts all paths from a JSON value.
// This is the same algorithm as in the paths package.
func extractPaths(prefix string, value any, counts map[string]int) {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			childPath := prefix + "." + key
			extractPaths(childPath, val, counts)
		}
	case []any:
		for _, item := range v {
			childPath := prefix + "[]"
			extractPaths(childPath, item, counts)
		}
	default:
		// Leaf value
		counts[prefix]++
	}
}

// extractPathsWithValues extracts paths and their values for distinct counting.
// Values are converted to strings for comparison.
//
// Python comparison:
//
//	def extract_with_values(prefix, value, result):
//	    if isinstance(value, dict):
//	        for k, v in value.items():
//	            extract_with_values(f"{prefix}.{k}", v, result)
//	    elif isinstance(value, list):
//	        for item in value:
//	            extract_with_values(f"{prefix}[]", item, result)
//	    else:
//	        result[prefix].append(str(value))
func extractPathsWithValues(prefix string, value any, pathValues map[string][]string) {
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			childPath := prefix + "." + key
			extractPathsWithValues(childPath, val, pathValues)
		}
	case []any:
		for _, item := range v {
			childPath := prefix + "[]"
			extractPathsWithValues(childPath, item, pathValues)
		}
	default:
		// Leaf value - convert to string for distinct counting
		strVal := valueToString(value)
		pathValues[prefix] = append(pathValues[prefix], strVal)
	}
}

// valueToString converts a JSON value to a string for distinct value comparison.
// Special values are displayed with angle-bracket labels for clarity.
func valueToString(v any) string {
	if v == nil {
		return "<null>"
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return "<empty>"
		}
		return val
	case float64:
		// Format numbers consistently
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// getTopValues returns the top N most frequent values from a frequency map.
func getTopValues(freqs map[string]int, n int) []ValueFrequency {
	if len(freqs) == 0 {
		return nil
	}

	// Convert to slice
	values := make([]ValueFrequency, 0, len(freqs))
	for val, count := range freqs {
		values = append(values, ValueFrequency{Value: val, Count: count})
	}

	// Sort by count descending, then value ascending
	sort.Slice(values, func(i, j int) bool {
		if values[i].Count != values[j].Count {
			return values[i].Count > values[j].Count
		}
		return values[i].Value < values[j].Value
	})

	// Return top N
	if len(values) > n {
		return values[:n]
	}
	return values
}
