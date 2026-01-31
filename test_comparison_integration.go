// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"jsondiff/internal/loganalyzer"
)

func main() {
	// Test files
	mainLog := "/private/tmp/claude/-Users-areese-projects-personal-jtool/b27b67ef-d3a4-4e33-be09-06b632a315a1/scratchpad/main.log"
	wipLog := "/private/tmp/claude/-Users-areese-projects-personal-jtool/b27b67ef-d3a4-4e33-be09-06b632a315a1/scratchpad/wip.log"

	// Analyze both files
	fmt.Println("Analyzing main.log...")
	leftResult, err := loganalyzer.AnalyzeFile(mainLog)
	if err != nil {
		panic(err)
	}

	fmt.Println("Analyzing wip.log...")
	rightResult, err := loganalyzer.AnalyzeFile(wipLog)
	if err != nil {
		panic(err)
	}

	// Compare
	fmt.Println("\nComparing results...")
	comparison := loganalyzer.CompareAnalyses(leftResult, rightResult, mainLog, wipLog)

	// Display results as JSON
	output, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("\nComparison Result:")
	fmt.Println(string(output))

	// Display summary
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total paths: %d\n", comparison.Stats.TotalPaths)
	fmt.Printf("Added: %d\n", comparison.Stats.AddedPaths)
	fmt.Printf("Removed: %d\n", comparison.Stats.RemovedPaths)
	fmt.Printf("Changed: %d\n", comparison.Stats.ChangedPaths)
	fmt.Printf("Equal: %d\n", comparison.Stats.EqualPaths)

	// Verify expectations
	fmt.Println("\n=== Verification ===")
	expectedAdded := 1    // .record.email
	expectedChanged := 4  // .type, .stream, .record.id, .record.name (count changed from 2 to 3)
	expectedRemoved := 0
	expectedEqual := 0

	if comparison.Stats.AddedPaths == expectedAdded {
		fmt.Println("✓ Added paths count matches")
	} else {
		fmt.Printf("✗ Expected %d added, got %d\n", expectedAdded, comparison.Stats.AddedPaths)
	}

	if comparison.Stats.ChangedPaths == expectedChanged {
		fmt.Println("✓ Changed paths count matches")
	} else {
		fmt.Printf("✗ Expected %d changed, got %d\n", expectedChanged, comparison.Stats.ChangedPaths)
	}

	if comparison.Stats.RemovedPaths == expectedRemoved {
		fmt.Println("✓ Removed paths count matches")
	} else {
		fmt.Printf("✗ Expected %d removed, got %d\n", expectedRemoved, comparison.Stats.RemovedPaths)
	}

	if comparison.Stats.EqualPaths == expectedEqual {
		fmt.Println("✓ Equal paths count matches")
	} else {
		fmt.Printf("✗ Expected %d equal, got %d\n", expectedEqual, comparison.Stats.EqualPaths)
	}
}
