# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development

```bash
make build          # go build -o ./bin/restclient-cli ./cmd/
make test           # go test -race -count=1 ./...
make vet            # go vet ./...
make fmt            # go fmt ./...
make lint           # golangci-lint run ./...
make run            # build + run with examples/jsonplaceholder.http
make install        # go install ./cmd/

# Run a single test
go test -race -run TestParseSingleGetRequest ./internal/parser/
```

Requires Go 1.26+.

## Architecture

```
cmd/main.go              # CLI entry point: parse args, parse .http file, launch TUI
internal/
  parser/parser.go       # .http file parser: variable substitution, block splitting, request extraction
  executor/executor.go   # HTTP request execution with error wrapping (timeout, DNS, connection refused)
  tui/
    model.go             # Bubble Tea model: 3-panel layout, keyboard handling, async request dispatch
    panels.go            # Rendering: request list items, detail view, response view, JSON pretty-print, help
    styles.go            # Lipgloss styles: method badges, status codes, panel borders, color scheme
    keys.go              # Key bindings (unused — keyboard handling is inline in model.go)
```

**Data flow:** `parser.ParseFile()` → `[]parser.Request` → `tui.NewModel()` wraps them in a Bubble Tea program → on Enter, `executor.Execute()` runs the HTTP request asynchronously via `tea.Cmd`.

**TUI panels:** Left column (30% width) = request list with filter. Right column = request detail (top) + response (bottom). Status bar at bottom. Tab cycles focus between the three panels.

**Variable substitution:** `@name = value` defines variables; `{{name}}` references them in URLs, headers, and bodies. Undefined variables produce parse errors but don't block other requests.

**HTTP execution:** 30s default timeout, max 10 redirects. Transport cloned from `http.DefaultTransport` per request (overridable via `transportFactory` var for tests). Errors classified into timeout, DNS failure, and connection refused.
