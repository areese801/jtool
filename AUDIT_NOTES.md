# Security/Code-Smell Notes (Triage Later)

## Summary of concerns

1) DOM XSS surface in the webview UI
- The diff UI renders `node.path` directly into `innerHTML`, and JSON keys can influence paths.
- In a desktop webview, this means arbitrary JS in the app context (can call Wails bindings, access app data, potentially exfiltrate if network is available).
- See `frontend/src/main.js:315`, `frontend/src/main.js:327`, `frontend/src/main.js:344`.

2) Error messages injected via `innerHTML`
- Error strings are inserted with `innerHTML` in multiple places.
- If an upstream error ever contains user-controlled content, this becomes an XSS sink.
- See `frontend/src/main.js:279`, `frontend/src/main.js:407`, `frontend/src/main.js:482`.

3) Unbounded JSON size/depth can cause DoS
- JSON inputs are fully unmarshaled and recursively walked without explicit size/depth limits.
- Risk: memory exhaustion or stack overflow with very large or deeply nested input.
- See `app.go:66`, `app.go:91`, `app.go:118`, `internal/diff/diff.go:83`, `internal/normalize/normalize.go:30`, `internal/paths/paths.go:87`, `internal/loganalyzer/analyzer.go:70`.
