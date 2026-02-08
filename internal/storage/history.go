// Package storage handles persistent storage for application data.
// This is similar to Python's configparser or json.dump/load for app state.
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// FileHistory stores the history of file paths for different inputs.
// Each key represents a different file input (e.g., "diff-left", "diff-right", "paths").
//
// Python equivalent:
//   history = {
//       "diff-left": ["path1", "path2", ...],
//       "diff-right": ["path1", "path2", ...],
//   }
type FileHistory struct {
	Paths map[string][]string `json:"paths"` // Key -> list of paths (newest first)
	mu    sync.RWMutex        `json:"-"`     // Mutex for thread-safe access (not serialized)
}

const (
	maxHistoryPerKey = 10           // Maximum number of paths to store per key
	historyFileName  = "history.json" // File name for storing history
)

// NewFileHistory creates a new FileHistory instance.
//
// Python equivalent:
//   def new_file_history():
//       return {"paths": {}}
func NewFileHistory() *FileHistory {
	return &FileHistory{
		Paths: make(map[string][]string),
	}
}

// Add adds a file path to the history for a specific key.
// If the path already exists, it's moved to the front (most recent).
// Keeps only the most recent maxHistoryPerKey paths.
//
// Python equivalent:
//   def add(history, key, path):
//       if key not in history["paths"]:
//           history["paths"][key] = []
//       paths = history["paths"][key]
//       if path in paths:
//           paths.remove(path)
//       paths.insert(0, path)
//       history["paths"][key] = paths[:10]
func (h *FileHistory) Add(key, path string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize the slice if it doesn't exist
	if h.Paths[key] == nil {
		h.Paths[key] = make([]string, 0, maxHistoryPerKey)
	}

	// Remove path if it already exists (we'll add it to the front)
	paths := h.Paths[key]
	for i, p := range paths {
		if p == path {
			// Remove by slicing around the element
			// Python: paths.remove(path)
			paths = append(paths[:i], paths[i+1:]...)
			break
		}
	}

	// Add to front
	// Python: paths.insert(0, path)
	paths = append([]string{path}, paths...)

	// Keep only the most recent maxHistoryPerKey paths
	// Python: paths[:10]
	if len(paths) > maxHistoryPerKey {
		paths = paths[:maxHistoryPerKey]
	}

	h.Paths[key] = paths
}

// Get returns the file path history for a specific key.
// Returns an empty slice if the key doesn't exist.
//
// Python equivalent:
//   def get(history, key):
//       return history["paths"].get(key, [])
func (h *FileHistory) Get(key string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	paths := h.Paths[key]
	if paths == nil {
		return []string{}
	}

	// Return a copy to prevent external modification
	// Python: return paths.copy()
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// Clear removes all file path history.
//
// Python equivalent:
//   def clear(history):
//       history["paths"] = {}
func (h *FileHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Paths = make(map[string][]string)
}

// GetMostRecent returns the most recent file path for a specific key.
// Returns an empty string if no history exists.
//
// Python equivalent:
//   def get_most_recent(history, key):
//       paths = history["paths"].get(key, [])
//       return paths[0] if paths else ""
func (h *FileHistory) GetMostRecent(key string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	paths := h.Paths[key]
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// Save writes the history to a JSON file.
// The file is stored in the user's config directory.
//
// Python equivalent:
//   import json
//   def save(history, filepath):
//       with open(filepath, 'w') as f:
//           json.dump(history, f, indent=2)
func (h *FileHistory) Save(configDir string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure config directory exists
	// Python: os.makedirs(config_dir, exist_ok=True)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(configDir, historyFileName)

	// Marshal to JSON
	// Python: json.dumps(history, indent=2)
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	// Python: with open(..., 'w') as f: f.write(data)
	return os.WriteFile(filePath, data, 0644)
}

// Load reads the history from a JSON file.
// If the file doesn't exist, returns an empty history (not an error).
//
// Python equivalent:
//   import json
//   def load(filepath):
//       try:
//           with open(filepath, 'r') as f:
//               return json.load(f)
//       except FileNotFoundError:
//           return {"paths": {}}
func Load(configDir string) (*FileHistory, error) {
	filePath := filepath.Join(configDir, historyFileName)

	// Check if file exists
	// Python: os.path.exists(filepath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return empty history
		return NewFileHistory(), nil
	}

	// Read file
	// Python: with open(filepath, 'r') as f: data = f.read()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	// Python: json.loads(data)
	var history FileHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	// Initialize the map if it's nil (shouldn't happen, but be safe)
	if history.Paths == nil {
		history.Paths = make(map[string][]string)
	}

	return &history, nil
}
