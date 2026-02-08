package loganalyzer

import (
	"testing"
)

func TestCompareAnalyses(t *testing.T) {
	// Table-driven tests - a Go idiom for comprehensive testing
	// Similar to pytest's parametrize decorator in Python
	tests := []struct {
		name      string
		left      *AnalysisResult
		right     *AnalysisResult
		leftFile  string
		rightFile string
		wantStats ComparisonStats
		checkFunc func(*testing.T, *ComparisonResult) // Custom validation
	}{
		{
			name: "identical results",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
					{Path: ".record.name", Count: 100, ObjectHits: 100, DistinctCount: 50},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
					{Path: ".record.name", Count: 100, ObjectHits: 100, DistinctCount: 50},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:   2,
				EqualPaths:   2,
				AddedPaths:   0,
				RemovedPaths: 0,
				ChangedPaths: 0,
			},
			checkFunc: func(t *testing.T, result *ComparisonResult) {
				// All comparisons should have status equal
				for _, comp := range result.Comparisons {
					if comp.Status != StatusEqual {
						t.Errorf("expected all paths to be equal, got %s for path %s", comp.Status, comp.Path)
					}
					if comp.CountDelta != 0 {
						t.Errorf("expected zero delta for %s, got %d", comp.Path, comp.CountDelta)
					}
				}
			},
		},
		{
			name: "new path added",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
					{Path: ".record.email", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         2,
				EqualPaths:         1,
				AddedPaths:         1,
				RemovedPaths:       0,
				ChangedPaths:       0,
				TotalCountDelta:    100,
				TotalObjectsDelta:  100,
				TotalDistinctDelta: 100,
			},
			checkFunc: func(t *testing.T, result *ComparisonResult) {
				// Find the added path
				var addedComp *PathComparison
				for i := range result.Comparisons {
					if result.Comparisons[i].Path == ".record.email" {
						addedComp = &result.Comparisons[i]
						break
					}
				}
				if addedComp == nil {
					t.Fatal("expected to find .record.email in comparisons")
				}
				if addedComp.Status != StatusAdded {
					t.Errorf("expected .record.email to be added, got %s", addedComp.Status)
				}
				if addedComp.Left != nil {
					t.Error("expected left to be nil for added path")
				}
				if addedComp.Right == nil {
					t.Error("expected right to be non-nil for added path")
				}
				if addedComp.CountDelta != 100 {
					t.Errorf("expected count delta of 100, got %d", addedComp.CountDelta)
				}
			},
		},
		{
			name: "path removed",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
					{Path: ".old_field", Count: 50, ObjectHits: 50, DistinctCount: 25},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         2,
				EqualPaths:         1,
				AddedPaths:         0,
				RemovedPaths:       1,
				ChangedPaths:       0,
				TotalCountDelta:    -50,
				TotalObjectsDelta:  -50,
				TotalDistinctDelta: -25,
			},
			checkFunc: func(t *testing.T, result *ComparisonResult) {
				// Find the removed path
				var removedComp *PathComparison
				for i := range result.Comparisons {
					if result.Comparisons[i].Path == ".old_field" {
						removedComp = &result.Comparisons[i]
						break
					}
				}
				if removedComp == nil {
					t.Fatal("expected to find .old_field in comparisons")
				}
				if removedComp.Status != StatusRemoved {
					t.Errorf("expected .old_field to be removed, got %s", removedComp.Status)
				}
				if removedComp.Left == nil {
					t.Error("expected left to be non-nil for removed path")
				}
				if removedComp.Right != nil {
					t.Error("expected right to be nil for removed path")
				}
				if removedComp.CountDelta != -50 {
					t.Errorf("expected count delta of -50, got %d", removedComp.CountDelta)
				}
			},
		},
		{
			name: "count changed",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 150, ObjectHits: 150, DistinctCount: 150},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         1,
				EqualPaths:         0,
				AddedPaths:         0,
				RemovedPaths:       0,
				ChangedPaths:       1,
				TotalCountDelta:    50,
				TotalObjectsDelta:  50,
				TotalDistinctDelta: 50,
			},
			checkFunc: func(t *testing.T, result *ComparisonResult) {
				if len(result.Comparisons) != 1 {
					t.Fatalf("expected 1 comparison, got %d", len(result.Comparisons))
				}
				comp := result.Comparisons[0]
				if comp.Status != StatusChanged {
					t.Errorf("expected changed status, got %s", comp.Status)
				}
				if comp.CountDelta != 50 {
					t.Errorf("expected count delta of 50, got %d", comp.CountDelta)
				}
				if comp.Left == nil || comp.Right == nil {
					t.Error("expected both left and right to be non-nil for changed path")
				}
			},
		},
		{
			name: "empty left result",
			left: &AnalysisResult{
				Paths: []PathSummary{},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         1,
				EqualPaths:         0,
				AddedPaths:         1,
				RemovedPaths:       0,
				ChangedPaths:       0,
				TotalCountDelta:    100,
				TotalObjectsDelta:  100,
				TotalDistinctDelta: 100,
			},
		},
		{
			name: "empty right result",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".record.id", Count: 100, ObjectHits: 100, DistinctCount: 100},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         1,
				EqualPaths:         0,
				AddedPaths:         0,
				RemovedPaths:       1,
				ChangedPaths:       0,
				TotalCountDelta:    -100,
				TotalObjectsDelta:  -100,
				TotalDistinctDelta: -100,
			},
		},
		{
			name: "both empty results",
			left: &AnalysisResult{
				Paths: []PathSummary{},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths: 0,
			},
		},
		{
			name:      "nil left input",
			left:      nil,
			right:     &AnalysisResult{},
			leftFile:  "",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths: 0,
			},
		},
		{
			name:      "nil right input",
			left:      &AnalysisResult{},
			right:     nil,
			leftFile:  "main.log",
			rightFile: "",
			wantStats: ComparisonStats{
				TotalPaths: 0,
			},
		},
		{
			name: "complex scenario with mixed changes",
			left: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".type", Count: 2, ObjectHits: 2, DistinctCount: 1},
					{Path: ".stream", Count: 2, ObjectHits: 2, DistinctCount: 1},
					{Path: ".record.id", Count: 2, ObjectHits: 2, DistinctCount: 2},
					{Path: ".record.name", Count: 2, ObjectHits: 2, DistinctCount: 2},
				},
			},
			right: &AnalysisResult{
				Paths: []PathSummary{
					{Path: ".type", Count: 3, ObjectHits: 3, DistinctCount: 1},
					{Path: ".stream", Count: 3, ObjectHits: 3, DistinctCount: 1},
					{Path: ".record.id", Count: 3, ObjectHits: 3, DistinctCount: 3},
					{Path: ".record.name", Count: 3, ObjectHits: 3, DistinctCount: 3},
					{Path: ".record.email", Count: 3, ObjectHits: 3, DistinctCount: 3},
				},
			},
			leftFile:  "main.log",
			rightFile: "wip.log",
			wantStats: ComparisonStats{
				TotalPaths:         5,
				EqualPaths:         0,
				AddedPaths:         1,
				RemovedPaths:       0,
				ChangedPaths:       4,
				TotalCountDelta:    7,
				TotalObjectsDelta:  7,
				TotalDistinctDelta: 5,
			},
		},
	}

	// Run all test cases
	// In Go, t.Run creates subtests - similar to pytest's test parametrization
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareAnalyses(tt.left, tt.right, tt.leftFile, tt.rightFile)

			// Check basic stats
			if result.Stats.TotalPaths != tt.wantStats.TotalPaths {
				t.Errorf("TotalPaths = %d, want %d", result.Stats.TotalPaths, tt.wantStats.TotalPaths)
			}
			if result.Stats.AddedPaths != tt.wantStats.AddedPaths {
				t.Errorf("AddedPaths = %d, want %d", result.Stats.AddedPaths, tt.wantStats.AddedPaths)
			}
			if result.Stats.RemovedPaths != tt.wantStats.RemovedPaths {
				t.Errorf("RemovedPaths = %d, want %d", result.Stats.RemovedPaths, tt.wantStats.RemovedPaths)
			}
			if result.Stats.ChangedPaths != tt.wantStats.ChangedPaths {
				t.Errorf("ChangedPaths = %d, want %d", result.Stats.ChangedPaths, tt.wantStats.ChangedPaths)
			}
			if result.Stats.EqualPaths != tt.wantStats.EqualPaths {
				t.Errorf("EqualPaths = %d, want %d", result.Stats.EqualPaths, tt.wantStats.EqualPaths)
			}
			if result.Stats.TotalCountDelta != tt.wantStats.TotalCountDelta {
				t.Errorf("TotalCountDelta = %d, want %d", result.Stats.TotalCountDelta, tt.wantStats.TotalCountDelta)
			}
			if result.Stats.TotalObjectsDelta != tt.wantStats.TotalObjectsDelta {
				t.Errorf("TotalObjectsDelta = %d, want %d", result.Stats.TotalObjectsDelta, tt.wantStats.TotalObjectsDelta)
			}
			if result.Stats.TotalDistinctDelta != tt.wantStats.TotalDistinctDelta {
				t.Errorf("TotalDistinctDelta = %d, want %d", result.Stats.TotalDistinctDelta, tt.wantStats.TotalDistinctDelta)
			}

			// Run custom validation if provided
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}

			// Verify file paths are preserved
			if result.LeftFile != tt.leftFile {
				t.Errorf("LeftFile = %s, want %s", result.LeftFile, tt.leftFile)
			}
			if result.RightFile != tt.rightFile {
				t.Errorf("RightFile = %s, want %s", result.RightFile, tt.rightFile)
			}
		})
	}
}

func TestStatsAreEqual(t *testing.T) {
	tests := []struct {
		name  string
		left  PathSummary
		right PathSummary
		want  bool
	}{
		{
			name:  "identical summaries",
			left:  PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			right: PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			want:  true,
		},
		{
			name:  "different count",
			left:  PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			right: PathSummary{Count: 150, ObjectHits: 100, DistinctCount: 50},
			want:  false,
		},
		{
			name:  "different object hits",
			left:  PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			right: PathSummary{Count: 100, ObjectHits: 150, DistinctCount: 50},
			want:  false,
		},
		{
			name:  "different distinct count",
			left:  PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			right: PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 75},
			want:  false,
		},
		{
			name:  "all different",
			left:  PathSummary{Count: 100, ObjectHits: 100, DistinctCount: 50},
			right: PathSummary{Count: 200, ObjectHits: 200, DistinctCount: 100},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statsAreEqual(tt.left, tt.right)
			if got != tt.want {
				t.Errorf("statsAreEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComparisonSorting(t *testing.T) {
	// Test that comparisons are sorted by priority
	left := &AnalysisResult{
		Paths: []PathSummary{
			{Path: ".equal", Count: 100, ObjectHits: 100, DistinctCount: 100},
			{Path: ".removed", Count: 50, ObjectHits: 50, DistinctCount: 50},
			{Path: ".changed", Count: 100, ObjectHits: 100, DistinctCount: 100},
		},
	}
	right := &AnalysisResult{
		Paths: []PathSummary{
			{Path: ".equal", Count: 100, ObjectHits: 100, DistinctCount: 100},
			{Path: ".added", Count: 75, ObjectHits: 75, DistinctCount: 75},
			{Path: ".changed", Count: 150, ObjectHits: 150, DistinctCount: 150},
		},
	}

	result := CompareAnalyses(left, right, "main.log", "wip.log")

	// Verify sorting order: removed, added, changed, equal
	expectedOrder := []struct {
		path   string
		status ComparisonStatus
	}{
		{".removed", StatusRemoved},
		{".added", StatusAdded},
		{".changed", StatusChanged},
		{".equal", StatusEqual},
	}

	if len(result.Comparisons) != len(expectedOrder) {
		t.Fatalf("expected %d comparisons, got %d", len(expectedOrder), len(result.Comparisons))
	}

	for i, expected := range expectedOrder {
		if result.Comparisons[i].Path != expected.path {
			t.Errorf("position %d: expected path %s, got %s", i, expected.path, result.Comparisons[i].Path)
		}
		if result.Comparisons[i].Status != expected.status {
			t.Errorf("position %d: expected status %s, got %s", i, expected.status, result.Comparisons[i].Status)
		}
	}
}
