package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"jsondiff/internal/diff"
	"jsondiff/internal/loganalyzer"
	"jsondiff/internal/normalize"
	"jsondiff/internal/paths"
)

// App struct holds the application state.
//
// Python comparison:
//   - Like a Python class with __init__ storing instance variables
//   - ctx is like storing request context in Flask/Django
//   - Go structs only hold data; methods are defined separately
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
// This is Go's equivalent of a constructor/factory function.
//
// Python comparison:
//
//	def __init__(self):
//	    pass  # No initialization needed yet
//
// In Go, we return a pointer (*App) so the caller gets a reference
// to the same object, not a copy.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
//
// This is a Wails lifecycle hook - called automatically when app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CompareJSON takes two JSON strings, parses them, and returns the diff result.
// This method is exposed to the frontend via Wails bindings.
//
// Python comparison:
//
//	def compare_json(self, left_json: str, right_json: str) -> DiffResult:
//	    try:
//	        left = json.loads(left_json)
//	        right = json.loads(right_json)
//	        return diff.compare(left, right)
//	    except json.JSONDecodeError as e:
//	        raise ValueError(f"Invalid JSON: {e}")
//
// Key Go differences:
//   - Returns (result, error) instead of raising exceptions
//   - Must explicitly check and handle errors
//   - The `any` type receives the parsed JSON (like Python's dynamic typing)
func (a *App) CompareJSON(leftJSON, rightJSON string) (*diff.DiffResult, error) {
	// Parse left JSON
	var left any
	if err := json.Unmarshal([]byte(leftJSON), &left); err != nil {
		return nil, fmt.Errorf("invalid left JSON: %w", err)
	}

	// Parse right JSON
	var right any
	if err := json.Unmarshal([]byte(rightJSON), &right); err != nil {
		return nil, fmt.Errorf("invalid right JSON: %w", err)
	}

	// Perform the diff
	result := diff.Compare(left, right)
	return result, nil
}

// FormatJSON takes a JSON string and returns it pretty-printed.
// Useful for normalizing user input in the UI.
//
// Python comparison:
//
//	def format_json(self, json_str: str) -> str:
//	    return json.dumps(json.loads(json_str), indent=2)
func (a *App) FormatJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// MarshalIndent pretty-prints with indentation
	// Like json.dumps(data, indent=2) in Python
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}

	return string(formatted), nil
}

// ValidateJSON checks if a string is valid JSON.
// Returns an error message if invalid, empty string if valid.
//
// Python comparison:
//
//	def validate_json(self, json_str: str) -> str:
//	    try:
//	        json.loads(json_str)
//	        return ""
//	    except json.JSONDecodeError as e:
//	        return str(e)
func (a *App) ValidateJSON(jsonStr string) string {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return err.Error()
	}
	return ""
}

// NormalizeOptions mirrors the normalize.Options struct for frontend use.
// Wails automatically converts between Go structs and JavaScript objects.
//
// Python comparison:
//   - Like a TypedDict or Pydantic model for API serialization
//   - Wails handles the JSON serialization automatically
type NormalizeOptions struct {
	SortKeys         bool   `json:"sortKeys"`
	NormalizeNumbers bool   `json:"normalizeNumbers"`
	TrimStrings      bool   `json:"trimStrings"`
	NullEqualsAbsent bool   `json:"nullEqualsAbsent"`
	SortArrays       bool   `json:"sortArrays"`
	SortArraysByKey  string `json:"sortArraysByKey"`
}

// CompareJSONWithOptions compares two JSON strings with normalization options.
// This is the "smart" comparison that handles key ordering, number formats, etc.
//
// Python comparison:
//
//	def compare_json_with_options(self, left: str, right: str, opts: dict) -> DiffResult:
//	    left_data = json.loads(left)
//	    right_data = json.loads(right)
//	    normalized_left = normalize(left_data, opts)
//	    normalized_right = normalize(right_data, opts)
//	    return diff.compare(normalized_left, normalized_right)
func (a *App) CompareJSONWithOptions(leftJSON, rightJSON string, opts NormalizeOptions) (*diff.DiffResult, error) {
	// Parse left JSON
	var left any
	if err := json.Unmarshal([]byte(leftJSON), &left); err != nil {
		return nil, fmt.Errorf("invalid left JSON: %w", err)
	}

	// Parse right JSON
	var right any
	if err := json.Unmarshal([]byte(rightJSON), &right); err != nil {
		return nil, fmt.Errorf("invalid right JSON: %w", err)
	}

	// Convert frontend options to internal options
	normalizeOpts := normalize.Options{
		SortKeys:         opts.SortKeys,
		NormalizeNumbers: opts.NormalizeNumbers,
		TrimStrings:      opts.TrimStrings,
		NullEqualsAbsent: opts.NullEqualsAbsent,
		SortArrays:       opts.SortArrays,
		SortArraysByKey:  opts.SortArraysByKey,
	}

	// Perform the diff with normalization
	result := diff.CompareWithOptions(left, right, normalizeOpts)
	return result, nil
}

// GetDefaultNormalizeOptions returns the default normalization options.
// Called by frontend to initialize the UI with sensible defaults.
func (a *App) GetDefaultNormalizeOptions() NormalizeOptions {
	defaults := normalize.DefaultOptions()
	return NormalizeOptions{
		SortKeys:         defaults.SortKeys,
		NormalizeNumbers: defaults.NormalizeNumbers,
		TrimStrings:      defaults.TrimStrings,
		NullEqualsAbsent: defaults.NullEqualsAbsent,
		SortArrays:       defaults.SortArrays,
		SortArraysByKey:  defaults.SortArraysByKey,
	}
}

// OpenJSONFile opens a file dialog for selecting a JSON file and returns its contents.
// Uses Wails' runtime.OpenFileDialog which is sandbox-compatible for Mac App Store.
//
// Python comparison:
//
//	def open_json_file(self) -> str:
//	    # In Python with tkinter:
//	    from tkinter import filedialog
//	    path = filedialog.askopenfilename(filetypes=[("JSON", "*.json")])
//	    if not path:
//	        return ""
//	    with open(path) as f:
//	        return f.read()
//
// Key differences:
//   - Wails handles the native file dialog (uses macOS Cocoa dialogs)
//   - The dialog is sandboxed - user must explicitly select the file
//   - Returns empty string if user cancels (no error)
func (a *App) OpenJSONFile() (string, error) {
	// Open file dialog with JSON filter
	// runtime.OpenFileDialog uses the app context we stored in startup()
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select JSON File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("error opening file dialog: %w", err)
	}

	// User cancelled - return empty string (not an error)
	if path == "" {
		return "", nil
	}

	// Read file contents
	// os.ReadFile is Go 1.16+ - reads entire file into memory
	// Python equivalent: open(path).read()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(data), nil
}

// GetJSONPaths extracts all JSON paths from a JSON string.
// Returns all paths to leaf values with occurrence counts.
// Useful for understanding the structure/schema of a JSON document.
//
// Python comparison (from get_json_paths.py):
//
//	def get_json_paths(input_file):
//	    j = json.loads(json_data)
//	    paths = defaultdict(int)
//	    for k in j.keys():
//	        unpack(pre_path="$.", test_obj=j[k], result_set=paths)
//	    return OrderedDict(sorted(paths.items()))
//
// Key Go differences:
//   - Type switch instead of isinstance() checks
//   - Explicit sorting (Go maps are unordered)
//   - Returns structured result instead of printing
func (a *App) GetJSONPaths(jsonStr string) (*paths.PathResult, error) {
	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract all paths
	result := paths.Extract(data)
	return result, nil
}

// AnalyzeLogFile opens a file dialog, reads the selected file, and analyzes
// all JSON lines within it. Non-JSON lines (like log messages) are skipped.
//
// This is useful for analyzing Singer tap output, JSONL files, or any log
// file with embedded JSON objects.
//
// Python comparison:
//
//	def analyze_log_file():
//	    path = filedialog.askopenfilename()
//	    paths = defaultdict(lambda: {"count": 0, "objects": 0})
//	    for line in open(path):
//	        try:
//	            obj = json.loads(line)
//	            for p in get_paths(obj):
//	                paths[p]["count"] += 1
//	        except json.JSONDecodeError:
//	            pass  # Skip non-JSON lines
//	    return paths
func (a *App) AnalyzeLogFile() (*loganalyzer.AnalysisResult, error) {
	// Open file dialog
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Log File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text/Log Files (*.txt, *.log, *.jsonl)",
				Pattern:     "*.txt;*.log;*.jsonl",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error opening file dialog: %w", err)
	}

	// User cancelled
	if path == "" {
		return nil, nil
	}

	// Analyze the file
	result, err := loganalyzer.AnalyzeFile(path)
	if err != nil {
		return nil, fmt.Errorf("error analyzing file: %w", err)
	}

	return result, nil
}

// AnalyzeLogString analyzes JSON lines from a string input.
// Useful for smaller inputs pasted directly into the UI.
func (a *App) AnalyzeLogString(content string) (*loganalyzer.AnalysisResult, error) {
	return loganalyzer.AnalyzeString(content)
}
