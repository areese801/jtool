# CLAUDE.md - AI Assistant Context

## Project Overview

This is a **JSON Diff Tool** - a high-performance GUI application for comparing and diffing JSON objects with intelligent key normalization (similar to Kaleidoscope's approach).

**Tech Stack:**
- Language: Go 1.22+
- GUI: Wails v2 (Go backend + web frontend)
- Frontend: Vanilla JS + CSS (or Svelte if needed)
- JSON Parsing: `github.com/goccy/go-json` or `github.com/bytedance/sonic`

## Developer Background

**IMPORTANT:** The developer has a Python background but is learning Go through this project.

**Assistant Instructions:**
- When implementing Go code, provide Python comparisons and contrasts where relevant
- Explain Go-specific concepts (goroutines, channels, interfaces, pointers, error handling, etc.) in relation to Python equivalents
- Highlight key differences: static typing, explicit error handling, no classes (structs + methods), memory management
- Point out Go idioms and best practices that differ from Python conventions
- When showing Go patterns, explain what the equivalent Python code would look like

### Example Comparisons to Make:
- **Error Handling:** Go's explicit `if err != nil` vs Python's try/except
- **Type System:** Go's static types vs Python's dynamic typing
- **Concurrency:** Goroutines/channels vs Python's threading/asyncio
- **Package Structure:** Go modules vs Python packages
- **Memory:** Go pointers and pass-by-value vs Python's reference semantics
- **Methods:** Go structs + methods vs Python classes
- **Nil/None:** How Go's nil compares to Python's None
- **Iteration:** Go's for loops vs Python's for/while/comprehensions

## Key Features

1. **Key Normalization** - Kaleidoscope-style semantic comparison
   - Sort object keys alphabetically
   - Optional array sorting (by key for object arrays)
   - Number normalization (1.0 == 1)
   - Configurable null/whitespace handling

2. **Structured Diff Algorithm** - Produces structured diff (not just text)
   - JSON path tracking (e.g., `$.users[0].name`)
   - Diff types: equal, added, removed, changed
   - Hierarchical diff tree with children

3. **Large File Handling** - Memory-efficient streaming for 100MB+ files
   - < 50MB: Load entirely
   - 50-500MB: Stream parse with index
   - > 500MB: Chunked comparison with progress

4. **Wails Integration** - Go backend methods exposed to JavaScript frontend

## Project Structure

```
jtool/
├── main.go                 # Wails app entry point
├── app.go                  # App struct with frontend-exposed methods
├── internal/
│   ├── diff/              # Core diff algorithm
│   ├── normalize/         # Key normalization logic
│   ├── parser/            # JSON parsing + streaming
│   └── model/             # Shared types
├── frontend/              # Web UI (HTML/JS/CSS)
└── testdata/              # Test fixtures
```

## Performance Targets

| File Size | Parse Time | Diff Time | Memory |
|-----------|------------|-----------|---------|
| 1 MB      | < 50ms     | < 100ms   | < 50MB  |
| 10 MB     | < 200ms    | < 500ms   | < 200MB |
| 100 MB    | < 2s       | < 5s      | < 1GB   |
| 500 MB    | < 10s      | < 30s     | < 2GB   |

## Development Commands

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Initialize project
wails init -n jtool -t vanilla

# Development mode (hot reload)
wails dev

# Build for production
wails build

# Run tests
go test ./internal/... -v

# Benchmark
go test ./internal/diff -bench=. -benchmem

# Large file tests
go test ./... -tags=largefile
```

## Release Workflow

The project uses GitHub Actions to build and release cross-platform binaries. The workflow is defined in `.github/workflows/release.yml`.

**Supported Platforms:**
- macOS arm64 (Apple Silicon)
- macOS amd64 (Intel)
- Windows amd64
- Linux amd64

**To create a release:**
```bash
# After merging to main, tag the release
git tag -a v1.0.0 -m "Release description"
git push origin v1.0.0
```

This triggers the workflow which:
1. Builds the Wails app on all 4 platforms
2. Packages each build (`.zip` for macOS/Windows, `.tar.gz` for Linux)
3. Creates a GitHub Release with auto-generated notes and attached binaries

**Version format:** Use semantic versioning (`vMAJOR.MINOR.PATCH`)

### Assistant Reminder

**After merging a branch to main:** Ask the user if they want to create a version tag and release. Remind them of the tagging commands above.

## Phased Development Approach

**CRITICAL**: This project follows a strict phased approach to balance learning Go with building functionality.

### Current Phase: **Phase 1 - MVP**

**Focus**: Learn Go basics and get a working diff tool

**In Scope:**
- Simple recursive diff algorithm (no normalization)
- Basic Wails UI with two text areas
- Load JSON strings and display diff results
- Files < 10MB only
- Standard library only (`encoding/json`)
- Unit tests with table-driven test pattern

**Out of Scope (Later Phases):**
- ❌ Key normalization (Phase 2)
- ❌ Array sorting (Phase 2)
- ❌ Large file handling / streaming (Phase 3)
- ❌ Performance optimization (Phase 3)
- ❌ File dialogs (Phase 4)
- ❌ Configuration files (Phase 4)
- ❌ Advanced UI features (Phase 4)

### Phase Overview

| Phase | Focus | Learning Objectives |
|-------|-------|---------------------|
| **1: MVP** (current) | Basic diff tool | Go types, error handling, testing, Wails basics |
| **2: Normalization** | Semantic comparison | Maps, slices, recursion, type assertions |
| **3: Performance** | Optimization | Benchmarking, profiling, streaming, goroutines |
| **4: Polish** | Production-ready | Cross-platform builds, sandboxing, App Store |

**Assistant Reminder**: When implementing features, always check which phase we're in. Politely defer features to later phases if the user requests something out of scope.

## Current Status

**Phase 1 - MVP**: In progress
- Documentation updated with phased approach
- Next: Initialize Wails project and implement core diff algorithm

See PROJECT_SPEC.md for complete specification and detailed phase breakdown.

## UI/UX Design Patterns

### Controls Layout Pattern

All tabs should follow a consistent controls layout:

1. **Controls positioned above content panels** - Options, buttons, and toggles appear at the top of the tab
2. **Left-aligned horizontal layout** - Use `class="controls controls-horizontal"` for left-aligned row layout
3. **Options in a panel** - Group related checkboxes/options in `class="options-panel"`
4. **Primary action button to the right** - The main action button (Compare, Extract, etc.) appears after options

**Example HTML structure:**
```html
<div class="controls controls-horizontal">
    <div class="options-panel">
        <label class="checkbox-label">
            <input type="checkbox" id="some-option">
            Option Label
        </label>
    </div>
    <button class="btn-primary" id="action-btn">Action</button>
</div>

<!-- Content panels below -->
<div class="editor-container">
    ...
</div>
```

**Key CSS classes:**
- `.controls` - Base flex container for controls
- `.controls-horizontal` - Makes it a left-aligned row (`flex-direction: row; justify-content: flex-start;`)
- `.options-panel` - Styled container for checkbox options
- `.btn-primary` - Primary action button styling

This pattern ensures visual consistency across all tabs (Diff, Path Explorer, Log Analyzer).

## Notes for AI Assistant

- Prioritize code clarity and learning over hyper-optimization initially
- Explain Go patterns as they're introduced
- Use this project as a teaching opportunity for Go best practices
- Reference Python equivalents to help bridge the knowledge gap
- **Follow the UI/UX Design Patterns above for consistent interface design**
