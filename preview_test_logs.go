//go:build ignore

package main

import (
	"fmt"
	"jtool/internal/loganalyzer"
)

func main() {
	left := "/private/tmp/claude/-Users-areese-projects-personal-jtool/b27b67ef-d3a4-4e33-be09-06b632a315a1/scratchpad/baseline.log"
	right := "/private/tmp/claude/-Users-areese-projects-personal-jtool/b27b67ef-d3a4-4e33-be09-06b632a315a1/scratchpad/comparison.log"

	leftResult, _ := loganalyzer.AnalyzeFile(left)
	rightResult, _ := loganalyzer.AnalyzeFile(right)
	comparison := loganalyzer.CompareAnalyses(leftResult, rightResult, left, right)

	fmt.Printf("=== STATISTICS ===\n")
	fmt.Printf("Total Paths: %d\n", comparison.Stats.TotalPaths)
	fmt.Printf("Added: %d\n", comparison.Stats.AddedPaths)
	fmt.Printf("Removed: %d\n", comparison.Stats.RemovedPaths)
	fmt.Printf("Changed: %d\n", comparison.Stats.ChangedPaths)
	fmt.Printf("Equal: %d\n\n", comparison.Stats.EqualPaths)

	fmt.Printf("=== ADDED PATHS (GREEN) ===\n")
	for _, c := range comparison.Comparisons {
		if c.Status == "added" {
			fmt.Printf("  + %s (count: %d)\n", c.Path, c.Right.Count)
		}
	}

	fmt.Printf("\n=== REMOVED PATHS (RED) ===\n")
	for _, c := range comparison.Comparisons {
		if c.Status == "removed" {
			fmt.Printf("  - %s (count: %d)\n", c.Path, c.Left.Count)
		}
	}

	fmt.Printf("\n=== CHANGED PATHS (ORANGE) ===\n")
	for _, c := range comparison.Comparisons {
		if c.Status == "changed" {
			fmt.Printf("  ~ %s (left: %d, right: %d, delta: %+d)\n",
				c.Path, c.Left.Count, c.Right.Count, c.CountDelta)
		}
	}

	// Print summary
	fmt.Printf("\n=== This is what you'll see in the UI ===\n")
	fmt.Printf("The table will show paths sorted by status (removed → added → changed → equal)\n")
	fmt.Printf("Each row will be color-coded: green (added), red (removed), orange (changed)\n")
}
