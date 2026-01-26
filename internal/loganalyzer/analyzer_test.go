package loganalyzer

import (
	"testing"
)

func TestAnalyzeString_JSONL(t *testing.T) {
	// Test single-line JSONL (existing behavior)
	jsonl := `{"name": "alice"}
{"name": "bob"}
{"name": "charlie"}`

	result, err := AnalyzeString(jsonl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSONLines != 3 {
		t.Errorf("expected 3 JSON objects, got %d", result.JSONLines)
	}
}

func TestAnalyzeString_MultiLineJSON(t *testing.T) {
	// Test multi-line pretty-printed JSON
	multiline := `{
  "name": "alice",
  "age": 30
}
{
  "name": "bob",
  "age": 25
}`

	result, err := AnalyzeString(multiline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSONLines != 2 {
		t.Errorf("expected 2 JSON objects, got %d", result.JSONLines)
	}

	// Verify paths were extracted
	foundName := false
	foundAge := false
	for _, p := range result.Paths {
		if p.Path == "$.name" {
			foundName = true
			if p.Count != 2 {
				t.Errorf("expected $.name count=2, got %d", p.Count)
			}
		}
		if p.Path == "$.age" {
			foundAge = true
			if p.Count != 2 {
				t.Errorf("expected $.age count=2, got %d", p.Count)
			}
		}
	}
	if !foundName {
		t.Error("$.name path not found")
	}
	if !foundAge {
		t.Error("$.age path not found")
	}
}

func TestAnalyzeString_MixedContent(t *testing.T) {
	// Test mixed content: logs + multi-line JSON
	mixed := `INFO: Starting process
{
  "type": "RECORD",
  "stream": "users"
}
DEBUG: Processing complete
{
  "type": "STATE",
  "value": {"position": 100}
}`

	result, err := AnalyzeString(mixed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSONLines != 2 {
		t.Errorf("expected 2 JSON objects, got %d", result.JSONLines)
	}

	// Verify both objects were parsed
	foundType := false
	for _, p := range result.Paths {
		if p.Path == "$.type" {
			foundType = true
			if p.Count != 2 {
				t.Errorf("expected $.type count=2, got %d", p.Count)
			}
		}
	}
	if !foundType {
		t.Error("$.type path not found")
	}
}

func TestAnalyzeString_NestedMultiLine(t *testing.T) {
	// Test deeply nested multi-line JSON
	nested := `{
  "users": [
    {
      "name": "alice",
      "profile": {
        "city": "NYC"
      }
    }
  ]
}`

	result, err := AnalyzeString(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSONLines != 1 {
		t.Errorf("expected 1 JSON object, got %d", result.JSONLines)
	}

	// Verify nested paths were extracted
	foundCity := false
	for _, p := range result.Paths {
		if p.Path == "$.users[].profile.city" {
			foundCity = true
		}
	}
	if !foundCity {
		t.Error("$.users[].profile.city path not found")
	}
}

func TestAnalyzeString_MultiLineArray(t *testing.T) {
	// Test multi-line JSON array
	array := `[
  1,
  2,
  3
]`

	result, err := AnalyzeString(array)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.JSONLines != 1 {
		t.Errorf("expected 1 JSON object (array), got %d", result.JSONLines)
	}
}
