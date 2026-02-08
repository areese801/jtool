package diff

// DiffType represents the type of difference found
type DiffType string

const (
	DiffEqual   DiffType = "equal"   // Values are the same
	DiffAdded   DiffType = "added"   // Value exists in right but not left
	DiffRemoved DiffType = "removed" // Value exists in left but not right
	DiffChanged DiffType = "changed" // Value exists in both but with different values
)

// DiffNode represents a single node in the diff tree
type DiffNode struct {
	Path     string     `json:"path"`               // JSON path (e.g., "$.users[0].name")
	Type     DiffType   `json:"type"`               // Type of difference
	Left     any        `json:"left,omitempty"`     // Value from left side (if applicable)
	Right    any        `json:"right,omitempty"`    // Value from right side (if applicable)
	Children []DiffNode `json:"children,omitempty"` // Nested differences
}

// DiffStats tracks statistics about the diff
type DiffStats struct {
	Added   int `json:"added"`   // Count of added values
	Removed int `json:"removed"` // Count of removed values
	Changed int `json:"changed"` // Count of changed values
	Equal   int `json:"equal"`   // Count of equal values
}

// DiffResult is the top-level result of a diff operation
type DiffResult struct {
	Root  DiffNode  `json:"root"`  // Root of the diff tree
	Stats DiffStats `json:"stats"` // Overall statistics
}
