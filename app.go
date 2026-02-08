package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"jtool/internal/diff"
	"jtool/internal/loganalyzer"
	"jtool/internal/normalize"
	"jtool/internal/paths"
	"jtool/internal/storage"
)

// App struct holds the application state.
type App struct {
	ctx       context.Context
	history   *storage.FileHistory
	configDir string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
// This is a Wails lifecycle hook - called automatically when app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Get user config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home dir
		homeDir = "."
	}
	a.configDir = filepath.Join(homeDir, ".jtool")

	// Load file path history from disk
	history, err := storage.Load(a.configDir)
	if err != nil {
		// If we can't load history, start with an empty one
		// This is not a fatal error - the app can still function
		history = storage.NewFileHistory()
	}
	a.history = history
}

// shutdown is called when the app is closing.
// Save the file history to disk.
func (a *App) shutdown(ctx context.Context) {
	// Save history to disk
	if a.history != nil {
		_ = a.history.Save(a.configDir)
	}
}

// CompareJSON takes two JSON strings, parses them, and returns the diff result.
// This method is exposed to the frontend via Wails bindings.
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
func (a *App) FormatJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// MarshalIndent pretty-prints with indentation
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}

	return string(formatted), nil
}

// ValidateJSON checks if a string is valid JSON.
// Returns an error message if invalid, empty string if valid.
func (a *App) ValidateJSON(jsonStr string) string {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return err.Error()
	}
	return ""
}

// NormalizeOptions mirrors the normalize.Options struct for frontend use.
// Wails automatically converts between Go structs and JavaScript objects.
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
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(data), nil
}

// GetJSONPaths extracts all JSON paths from a JSON string.
// Returns all paths to leaf values with occurrence counts.
// Useful for understanding the structure/schema of a JSON document.
func (a *App) GetJSONPaths(jsonStr string) (*paths.PathResult, error) {
	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract all paths (leaf values only)
	result := paths.Extract(data)
	return result, nil
}

// GetJSONPathsWithContainers extracts all JSON paths including container paths (objects/arrays).
// This is useful for seeing the full structure including intermediate objects.
func (a *App) GetJSONPathsWithContainers(jsonStr string, includeContainers bool) (*paths.PathResult, error) {
	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extract paths with options
	result := paths.ExtractWithOptions(data, paths.ExtractOptions{
		IncludeContainers: includeContainers,
	})
	return result, nil
}

// AnalyzeLogFile opens a file dialog, reads the selected file, and analyzes
// all JSON lines within it. Non-JSON lines (like log messages) are skipped.
//
// This is useful for analyzing Singer tap output, JSONL files, or any log
// file with embedded JSON objects.
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

// AnalyzeLogFilePath analyzes a log file at the given path.
// Unlike AnalyzeLogFile, this doesn't open a file dialog - it uses the provided path directly.
// Returns the path along with the analysis result so the frontend can display it.
func (a *App) AnalyzeLogFilePath(path string) (*loganalyzer.AnalysisResult, error) {
	if path == "" {
		return nil, fmt.Errorf("no file path provided")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Analyze the file
	result, err := loganalyzer.AnalyzeFile(path)
	if err != nil {
		return nil, fmt.Errorf("error analyzing file: %w", err)
	}

	return result, nil
}

// SelectAndAnalyzeLogFile opens a file dialog and returns both the path and analysis result.
// This replaces AnalyzeLogFile when the frontend needs to know the selected path.
func (a *App) SelectAndAnalyzeLogFile() (*LogFileResult, error) {
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

	return &LogFileResult{
		Path:   path,
		Result: result,
	}, nil
}

// LogFileResult combines the file path with analysis results.
type LogFileResult struct {
	Path   string                      `json:"path"`
	Result *loganalyzer.AnalysisResult `json:"result"`
}

// OpenJSONFileWithPath opens a file dialog and returns both path and contents.
func (a *App) OpenJSONFileWithPath() (*FileResult, error) {
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
		return nil, fmt.Errorf("error opening file dialog: %w", err)
	}

	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return &FileResult{
		Path:    path,
		Content: string(data),
	}, nil
}

// ReadFilePath reads a file from a given path.
func (a *App) ReadFilePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("no file path provided")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(data), nil
}

// FileResult combines a file path with its contents.
type FileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// CompareLogAnalyses compares two log analysis results and returns a structured comparison.
// This is the core comparison method used by other comparison functions.
func (a *App) CompareLogAnalyses(left, right *loganalyzer.AnalysisResult, leftFile, rightFile string) *loganalyzer.ComparisonResult {
	return loganalyzer.CompareAnalyses(left, right, leftFile, rightFile)
}

// CompareLogFiles analyzes and compares two log files at the given paths.
// This is a convenience method that combines file analysis and comparison.
func (a *App) CompareLogFiles(leftPath, rightPath string) (*loganalyzer.ComparisonResult, error) {
	// Validate inputs
	if leftPath == "" || rightPath == "" {
		return nil, fmt.Errorf("both file paths are required")
	}

	// Analyze left file
	leftResult, err := loganalyzer.AnalyzeFile(leftPath)
	if err != nil {
		return nil, fmt.Errorf("error analyzing left file: %w", err)
	}

	// Analyze right file
	rightResult, err := loganalyzer.AnalyzeFile(rightPath)
	if err != nil {
		return nil, fmt.Errorf("error analyzing right file: %w", err)
	}

	// Compare the results
	comparison := loganalyzer.CompareAnalyses(leftResult, rightResult, leftPath, rightPath)
	return comparison, nil
}

// SelectAndCompareLogFiles opens two file dialogs (left/baseline and right/comparison)
// and returns the comparison result. This is the main entry point for the compare mode UI.
func (a *App) SelectAndCompareLogFiles() (*loganalyzer.ComparisonResult, error) {
	// Open first dialog for left/baseline file
	leftPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Baseline Log File (Left)",
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
		return nil, fmt.Errorf("error opening left file dialog: %w", err)
	}

	// User cancelled first dialog
	if leftPath == "" {
		return nil, nil
	}

	// Open second dialog for right/comparison file
	rightPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Comparison Log File (Right)",
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
		return nil, fmt.Errorf("error opening right file dialog: %w", err)
	}

	// User cancelled second dialog
	if rightPath == "" {
		return nil, nil
	}

	// Analyze and compare the files
	return a.CompareLogFiles(leftPath, rightPath)
}

// ============================================================
// File Path History Methods
// ============================================================

// GetFileHistory returns the file path history for a specific key.
// Keys are: "diff-left", "diff-right", "paths", "logs", "compare-left", "compare-right"
func (a *App) GetFileHistory(key string) []string {
	if a.history == nil {
		return []string{}
	}
	return a.history.Get(key)
}

// GetMostRecentFilePath returns the most recent file path for a specific key.
// Returns empty string if no history exists.
func (a *App) GetMostRecentFilePath(key string) string {
	if a.history == nil {
		return ""
	}
	return a.history.GetMostRecent(key)
}

// SaveFilePathToHistory adds a file path to the history for a specific key.
// This is called automatically when files are loaded, but can also be called manually.
func (a *App) SaveFilePathToHistory(key, path string) error {
	if a.history == nil {
		return fmt.Errorf("history not initialized")
	}

	a.history.Add(key, path)

	// Save to disk immediately
	return a.history.Save(a.configDir)
}

// GetAllFileHistory returns the entire file history map.
// This is useful for initializing all dropdowns on app startup.
func (a *App) GetAllFileHistory() map[string][]string {
	if a.history == nil {
		return map[string][]string{}
	}

	// Build a copy of all history
	result := make(map[string][]string)
	keys := []string{
		"diff-left",
		"diff-right",
		"paths",
		"logs",
		"compare-left",
		"compare-right",
	}

	for _, key := range keys {
		result[key] = a.history.Get(key)
	}

	return result
}

// ClearFileHistory clears all file path history and saves the empty state.
func (a *App) ClearFileHistory() error {
	if a.history == nil {
		return nil
	}

	a.history.Clear()

	// Save the empty history
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appConfigDir := filepath.Join(configDir, "jtool")
	return a.history.Save(appConfigDir)
}

// ShowSettingsTab emits an event to the frontend to switch to the Settings tab.
// This is called from the Help menu in the application menu bar (works on all platforms).
func (a *App) ShowSettingsTab() {
	runtime.EventsEmit(a.ctx, "switchTab", "settings")
}
