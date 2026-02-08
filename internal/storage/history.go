// Package storage handles persistent storage for application data.
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// FileHistory stores the history of file paths for different inputs.
// Each key represents a different file input (e.g., "diff-left", "diff-right", "paths").
type FileHistory struct {
	Paths map[string][]string `json:"paths"` // Key -> list of paths (newest first)
	mu    sync.RWMutex        `json:"-"`     // Mutex for thread-safe access (not serialized)
}

const (
	maxHistoryPerKey = 10           // Maximum number of paths to store per key
	historyFileName  = "history.json" // File name for storing history
)

// NewFileHistory creates a new FileHistory instance.
func NewFileHistory() *FileHistory {
	return &FileHistory{
		Paths: make(map[string][]string),
	}
}

// Add adds a file path to the history for a specific key.
// If the path already exists, it's moved to the front (most recent).
// Keeps only the most recent maxHistoryPerKey paths.
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
			paths = append(paths[:i], paths[i+1:]...)
			break
		}
	}

	// Add to front
	paths = append([]string{path}, paths...)

	// Keep only the most recent maxHistoryPerKey paths
	if len(paths) > maxHistoryPerKey {
		paths = paths[:maxHistoryPerKey]
	}

	h.Paths[key] = paths
}

// Get returns the file path history for a specific key.
// Returns an empty slice if the key doesn't exist.
func (h *FileHistory) Get(key string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	paths := h.Paths[key]
	if paths == nil {
		return []string{}
	}

	// Return a copy to prevent external modification
	result := make([]string, len(paths))
	copy(result, paths)
	return result
}

// Clear removes all file path history.
func (h *FileHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Paths = make(map[string][]string)
}

// GetMostRecent returns the most recent file path for a specific key.
// Returns an empty string if no history exists.
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
func (h *FileHistory) Save(configDir string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(configDir, historyFileName)

	// Marshal to JSON
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, data, 0644)
}

// Load reads the history from a JSON file.
// If the file doesn't exist, returns an empty history (not an error).
func Load(configDir string) (*FileHistory, error) {
	filePath := filepath.Join(configDir, historyFileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return empty history
		return NewFileHistory(), nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
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
