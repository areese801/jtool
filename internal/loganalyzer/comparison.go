package loganalyzer

import (
	"sort"
)

// ComparisonStatus represents the comparison state of a JSON path
type ComparisonStatus string

const (
	StatusEqual   ComparisonStatus = "equal"
	StatusAdded   ComparisonStatus = "added"
	StatusRemoved ComparisonStatus = "removed"
	StatusChanged ComparisonStatus = "changed"
)

// PathComparison represents the comparison of a single JSON path between two analyses
type PathComparison struct {
	Path          string            `json:"path"`
	Status        ComparisonStatus  `json:"status"`
	Left          *PathSummary      `json:"left,omitempty"`
	Right         *PathSummary      `json:"right,omitempty"`
	CountDelta    int               `json:"countDelta"`
	ObjectsDelta  int               `json:"objectsDelta"`
	DistinctDelta int               `json:"distinctDelta"`
}

// ComparisonStats aggregates statistics across all path comparisons
type ComparisonStats struct {
	TotalPaths         int `json:"totalPaths"`
	AddedPaths         int `json:"addedPaths"`
	RemovedPaths       int `json:"removedPaths"`
	ChangedPaths       int `json:"changedPaths"`
	EqualPaths         int `json:"equalPaths"`
	TotalCountDelta    int `json:"totalCountDelta"`
	TotalObjectsDelta  int `json:"totalObjectsDelta"`
	TotalDistinctDelta int `json:"totalDistinctDelta"`
}

// ComparisonResult represents the overall comparison between two log analyses
type ComparisonResult struct {
	Comparisons []PathComparison `json:"comparisons"`
	Stats       ComparisonStats  `json:"stats"`
	LeftFile    string           `json:"leftFile"`
	RightFile   string           `json:"rightFile"`
}

// CompareAnalyses compares two analysis results and returns a structured comparison
// This is more efficient than using the generic JSON diff for flat list comparisons
// Time complexity: O(n log n), Space complexity: O(n)
//
// leftFile and rightFile are optional file paths for display purposes
func CompareAnalyses(left, right *AnalysisResult, leftFile, rightFile string) *ComparisonResult {
	if left == nil || right == nil {
		return &ComparisonResult{
			Comparisons: []PathComparison{},
			Stats:       ComparisonStats{},
			LeftFile:    leftFile,
			RightFile:   rightFile,
		}
	}

	// Build maps for O(1) lookup (like Python dict)
	// In Go, map keys are compared by value, strings are compared efficiently
	leftMap := make(map[string]PathSummary)
	rightMap := make(map[string]PathSummary)

	for _, summary := range left.Paths {
		leftMap[summary.Path] = summary
	}
	for _, summary := range right.Paths {
		rightMap[summary.Path] = summary
	}

	// Collect all unique paths from both results
	// Using a map as a set (like Python's set() type)
	allPaths := make(map[string]bool)
	for path := range leftMap {
		allPaths[path] = true
	}
	for path := range rightMap {
		allPaths[path] = true
	}

	// Convert to sorted slice for consistent ordering
	paths := make([]string, 0, len(allPaths))
	for path := range allPaths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Compare each path
	comparisons := make([]PathComparison, 0, len(paths))
	stats := ComparisonStats{}

	for _, path := range paths {
		leftSummary, inLeft := leftMap[path]
		rightSummary, inRight := rightMap[path]

		comparison := PathComparison{
			Path: path,
		}

		// Determine status and populate fields
		if inLeft && inRight {
			// Path exists in both - compare statistics
			comparison.Left = &leftSummary
			comparison.Right = &rightSummary
			comparison.CountDelta = rightSummary.Count - leftSummary.Count
			comparison.ObjectsDelta = rightSummary.ObjectHits - leftSummary.ObjectHits
			comparison.DistinctDelta = rightSummary.DistinctCount - leftSummary.DistinctCount

			if statsAreEqual(leftSummary, rightSummary) {
				comparison.Status = StatusEqual
				stats.EqualPaths++
			} else {
				comparison.Status = StatusChanged
				stats.ChangedPaths++
			}
		} else if inLeft && !inRight {
			// Path only in left - removed
			comparison.Status = StatusRemoved
			comparison.Left = &leftSummary
			comparison.CountDelta = -leftSummary.Count
			comparison.ObjectsDelta = -leftSummary.ObjectHits
			comparison.DistinctDelta = -leftSummary.DistinctCount
			stats.RemovedPaths++
		} else {
			// Path only in right - added
			comparison.Status = StatusAdded
			comparison.Right = &rightSummary
			comparison.CountDelta = rightSummary.Count
			comparison.ObjectsDelta = rightSummary.ObjectHits
			comparison.DistinctDelta = rightSummary.DistinctCount
			stats.AddedPaths++
		}

		comparisons = append(comparisons, comparison)

		// Aggregate deltas
		stats.TotalCountDelta += comparison.CountDelta
		stats.TotalObjectsDelta += comparison.ObjectsDelta
		stats.TotalDistinctDelta += comparison.DistinctDelta
	}

	stats.TotalPaths = len(comparisons)

	// Sort comparisons by priority: removed, added, changed, equal
	// Within each status, sort by absolute delta magnitude (descending)
	sort.Slice(comparisons, func(i, j int) bool {
		// Primary sort: status priority
		iPriority := statusPriority(comparisons[i].Status)
		jPriority := statusPriority(comparisons[j].Status)
		if iPriority != jPriority {
			return iPriority < jPriority
		}

		// Secondary sort: delta magnitude (descending)
		iDelta := abs(comparisons[i].CountDelta)
		jDelta := abs(comparisons[j].CountDelta)
		if iDelta != jDelta {
			return iDelta > jDelta
		}

		// Tertiary sort: path name (alphabetical)
		return comparisons[i].Path < comparisons[j].Path
	})

	return &ComparisonResult{
		Comparisons: comparisons,
		Stats:       stats,
		LeftFile:    leftFile,
		RightFile:   rightFile,
	}
}

// statsAreEqual checks if two PathSummary objects have identical statistics
// In Python, you might use __eq__ or dataclasses with frozen=True
func statsAreEqual(left, right PathSummary) bool {
	return left.Count == right.Count &&
		left.ObjectHits == right.ObjectHits &&
		left.DistinctCount == right.DistinctCount
}

// statusPriority returns sort priority for comparison status
// Lower numbers sort first
func statusPriority(status ComparisonStatus) int {
	switch status {
	case StatusRemoved:
		return 0
	case StatusAdded:
		return 1
	case StatusChanged:
		return 2
	case StatusEqual:
		return 3
	default:
		return 4
	}
}

// abs returns absolute value of an integer
// Go doesn't have a built-in abs for int (only float64)
// In Python you'd use abs() which works on any numeric type
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
