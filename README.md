# jtool

A fast, cross-platform JSON diff and analysis tool with intelligent normalization.

Built with [Go](https://go.dev/) and [Wails](https://wails.io/) for native performance on macOS, Windows, and Linux.

## Features

### JSON Diff
Compare two JSON documents with semantic understanding:

- **Sort Keys** - Ignore key ordering differences (`{"b":1,"a":2}` equals `{"a":2,"b":1}`)
- **Normalize Numbers** - Treat `1.0` and `1` as equal
- **Trim Strings** - Ignore leading/trailing whitespace in string values
- **Null = Absent** - Treat `{"key": null}` as equivalent to missing key

Two view modes:
- **Structured View** - Hierarchical tree showing exact paths of differences
- **Side-by-Side View** - Traditional two-column comparison

### Path Explorer
Extract and explore all JSON paths from a document:
- See every unique path in your JSON structure
- Count occurrences of each path
- Useful for understanding complex or unfamiliar JSON schemas

### Log Analyzer
Analyze JSON-lines log files (JSONL, Singer taps, etc.):

**Single File Analysis:**
- Parse log files containing JSON objects (one per line)
- Extract all unique paths across all objects
- See path frequency and distinct value counts
- Click any path to copy a `jq` command for extraction

**Compare Files:**
- Compare path structures between two log files
- Identify added/removed/changed paths
- Useful for comparing API responses, data pipeline outputs, etc.

## Installation

### Download

Download the latest release for your platform from the [Releases](https://github.com/areese801/jtool/releases) page:

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `jtool-darwin-arm64.zip` |
| macOS (Intel) | `jtool-darwin-amd64.zip` |
| Windows | `jtool-windows-amd64.zip` |
| Linux | `jtool-linux-amd64.tar.gz` |

### macOS
```bash
# Extract and move to Applications
unzip jtool-darwin-*.zip
mv jtool.app /Applications/
```

> **Note:** On first launch, you may need to right-click and select "Open" to bypass Gatekeeper, or go to System Preferences → Security & Privacy to allow the app.

### Windows
Extract the zip and run `jtool.exe`.

### Linux
```bash
tar -xzf jtool-linux-amd64.tar.gz
./jtool
```

## Usage

### Diff Tab

1. Paste or load JSON into the left and right panels
2. Configure normalization options (checkboxes above the panels)
3. Click **Compare** or the comparison runs automatically when both sides have valid JSON
4. View results in Structured or Side-by-Side mode

**Loading Files:**
- Click **Load** to open a file picker
- Or paste/type a file path and press **Enter**
- Use **Reload** to refresh from disk after external changes

### Path Explorer Tab

1. Paste or load JSON into the input panel
2. Toggle **Include containers** to show object/array paths (not just leaf values)
3. Click **Extract Paths**
4. Browse the table of all unique paths and their occurrence counts

### Log Analyzer Tab

**Single Analysis:**
1. Click **Load File** or enter a path to a log file
2. View extracted paths with counts, object hits, and distinct values
3. Click the distinct count to see top values for any path
4. Click any path to copy a `jq` extraction command

**Compare Files:**
1. Switch to **Compare Files** mode
2. Load a baseline (left) and comparison (right) file
3. Click **Compare**
4. Review added, removed, and changed paths between files

## Building from Source

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- [Node.js 18+](https://nodejs.org/) (for frontend build)

### Install Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Build

```bash
# Clone the repository
git clone https://github.com/areese801/jtool.git
cd jtool

# Build for your current platform
wails build

# The app will be in build/bin/
```

### Development

```bash
# Run in development mode with hot reload
wails dev

# Run tests
go test ./internal/... -v

# Run benchmarks
go test ./internal/diff -bench=. -benchmem
```

## Project Structure

```
jtool/
├── main.go                 # Wails app entry point
├── app.go                  # App struct with frontend-exposed methods
├── internal/
│   ├── diff/              # Core diff algorithm
│   ├── normalize/         # Key normalization logic
│   ├── loganalyzer/       # Log file analysis
│   ├── paths/             # JSON path extraction
│   └── storage/           # File history persistence
├── frontend/
│   ├── index.html         # Main HTML
│   └── src/
│       ├── main.js        # Frontend logic
│       └── style.css      # Styling
└── build/                 # Build assets and output
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Reporting Bugs

Found a bug? [Open an issue](https://github.com/areese801/jtool/issues/new?labels=bug) on GitHub.

You can also report bugs from within the app:
- **Help menu → Report a Bug...**
- **Settings tab → Feedback → Report a Bug**

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Wails](https://wails.io/) - Build desktop apps with Go and web technologies
- Inspired by [Kaleidoscope](https://kaleidoscope.app/) and similar diff tools
