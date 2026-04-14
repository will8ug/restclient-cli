## Context

This is a greenfield project. No existing codebase, no legacy constraints. The goal is an interactive TUI tool that understands the `.http` file format from VS Code's REST Client extension and lets users browse, select, and execute HTTP requests in a multi-panel terminal interface.

The `.http` format is a de facto standard across VS Code REST Client and JetBrains HTTP Client. It defines HTTP requests in plain text files with a simple, human-readable syntax. Currently, executing these files requires an IDE — there's no good standalone terminal option with an interactive UI.

## Goals / Non-Goals

**Goals:**
- Parse `.http` files compatible with VS Code REST Client's core syntax
- Present requests in an interactive TUI with keyboard navigation
- Execute HTTP requests asynchronously and display formatted responses in-place
- Support file-level variables and substitution
- Run as a single binary with no runtime dependencies
- Keep the initial scope minimal — cover the 80% use case with a polished experience

**Non-Goals:**
- In-TUI request editing / body composer (future — initially read-only from `.http` files)
- Environment files (`.env`) or multi-environment switching (future)
- OAuth / authentication helpers (future)
- Request history or cookie jar persistence (future)
- Response scripting or assertions (future)
- GraphQL-specific support (future)
- Certificate / mTLS configuration (future)
- Request chaining (using response values in subsequent requests) (future)
- Mouse support (future — keyboard-first for v1)

## Decisions

### 1. Language: Go

**Choice**: Go

**Rationale**: Go produces a single static binary with no runtime dependencies, has excellent standard library HTTP support (`net/http`), fast compilation, and cross-platform builds are trivial (`GOOS`/`GOARCH`). The TUI ecosystem in Go is best-in-class.

**Alternatives considered**:
- **Rust + Ratatui**: Excellent performance but slower development iteration; Go's Bubble Tea ecosystem is more mature for rapid TUI development
- **TypeScript/Node + Ink**: Requires Node.js runtime; defeats the "single binary" goal
- **Python + Textual**: Requires Python runtime; distribution is messy

### 2. TUI Framework: Bubble Tea + Ecosystem

**Choice**: `github.com/charmbracelet/bubbletea` with Bubbles, Lip Gloss

**Rationale**:
- **Elm Architecture** (Model-Update-View) provides clean separation of state, logic, and rendering
- **`tea.Cmd`** for async I/O — HTTP requests are fire-and-forget commands that return messages on completion, no manual goroutine management
- **Bubbles** component library provides text input, viewport (scrollable content), list, spinner, and table — all needed for a REST client
- **Lip Gloss** for CSS-like terminal styling (borders, padding, colors, alignment)
- Most popular Go TUI framework (~28k stars), actively maintained, largest ecosystem
- **Testable**: `Update()` is a pure function — easy to unit test state transitions

**Alternatives considered**:
- **tview**: Rich built-in widgets (Flex, Grid, Table) make layout faster, but state lives inside widgets making testing harder. Async requires manual `QueueUpdateDraw` goroutine management. Powers k9s but architecturally messier for async-heavy apps.
- **tcell**: Too low-level — we'd be building a widget framework from scratch
- **gocui**: Less active (last release ~8 years ago), smaller ecosystem

### 3. HTTP Client: net/http (standard library)

**Choice**: Go's built-in `net/http`

**Rationale**: Covers all HTTP methods, custom headers, request bodies, and timeouts. No need for a third-party HTTP library for basic functionality.

### 4. Architecture: Bubble Tea + Three-Layer Core

```
┌─────────────────────────────────────────────────┐
│                 Bubble Tea App                   │
│  ┌───────────┐  ┌───────────┐  ┌─────────────┐  │
│  │  Request   │  │  Request  │  │  Response   │  │
│  │   List     │  │  Detail   │  │  Viewer     │  │
│  │ (bubbles/  │  │ (viewport)│  │ (viewport)  │  │
│  │   list)    │  │           │  │             │  │
│  └───────────┘  └───────────┘  └─────────────┘  │
│                                                  │
│  Model → Update(msg) → View() → render string   │
└──────────────┬───────────────────────────────────┘
               │ tea.Cmd (async)
    ┌──────────┴──────────┐
    │                     │
┌───▼──────┐    ┌────────▼────────┐
│  Parser  │    │    Executor     │
│(.http→[]Request)│  (Request→Response)│
└──────────┘    └─────────────────┘
```

**Core layers** (unchanged, UI-agnostic):
- **Parser**: Reads `.http` file, produces a list of `Request` structs with resolved variables
- **Executor**: Takes a `Request`, makes the HTTP call via `tea.Cmd`, returns a `Response` as a `tea.Msg`

**TUI layer**:
- **App Model**: Holds all state — parsed requests, selected index, current response, loading state, active panel
- **Update**: Handles keyboard events, parser results, HTTP response messages, and panel switching
- **View**: Renders three-panel layout using Lip Gloss — request list (left), request detail (top-right), response viewer (bottom-right)

### 5. TUI Layout

```
┌─ Requests ──────┬─ Request Detail ───────────────┐
│ ▸ 1. Get users  │ GET https://api.example.com/    │
│   2. Create user│ Content-Type: application/json  │
│   3. Delete user│                                 │
│                 │ {"name": "test"}                │
│                 ├─ Response ──────────────────────┤
│                 │ 200 OK  (145ms)                 │
│                 │ Content-Type: application/json   │
│                 │                                 │
│                 │ {                               │
│                 │   "id": 1,                      │
│                 │   "name": "test"                │
│                 │ }                               │
├─────────────────┴────────────────────────────────┤
│ ↑↓ navigate  enter send  tab switch  q quit      │
└──────────────────────────────────────────────────┘
```

- **Left panel**: Request list from `.http` file (bubbles/list)
- **Top-right panel**: Selected request's detail (method, URL, headers, body) — read-only viewport
- **Bottom-right panel**: Response after execution (status, headers, body) — scrollable viewport
- **Status bar**: Keyboard shortcuts and current state

### 6. Keyboard Bindings

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate request list |
| `Enter` | Execute selected request |
| `Tab` | Cycle focus between panels |
| `q` / `Ctrl+C` | Quit |
| `/` | Filter requests by name |
| `?` | Toggle help overlay |

### 7. .http File Format Support (initial scope)

Supported syntax (unchanged from parser spec):
- Request separator: `###` (with optional description after)
- Request line: `METHOD URL HTTP/version` (HTTP version optional)
- Headers: `Header-Name: value` (one per line, immediately after request line)
- Body: everything after the first blank line until next `###` or EOF
- Comments: lines starting with `#` or `//`
- Variables: `@name = value` at file level, referenced as `{{name}}`

### 8. CLI Entry Point

The binary still accepts a file path argument: `restclient api.http`. This parses the file and launches the TUI. Error cases (file not found, parse errors) are displayed before entering the TUI or as error screens within it.

## Risks / Trade-offs

- **Format compatibility**: The `.http` format isn't formally specified. Edge cases in VS Code REST Client's parser may differ from ours. → **Mitigation**: Start with the well-documented core syntax; add compatibility tests against real-world `.http` files over time.
- **Single binary size**: Go + Bubble Tea binaries are ~10-15MB. → **Acceptable**: Still a single file, trivial to distribute.
- **No streaming**: Initial implementation reads entire response into memory. → **Mitigation**: Fine for API testing; add streaming support later if needed for large responses.
- **Terminal compatibility**: Bubble Tea relies on terminal capabilities (alternate screen, colors). Some minimal terminals may not render correctly. → **Mitigation**: Bubble Tea handles fallback gracefully; test on common terminals (iTerm2, Terminal.app, Windows Terminal, Alacritty).
- **No in-TUI editing**: Users must edit `.http` files externally and reload. → **Acceptable for v1**: Keeps scope tight. File watching or hot-reload could be added later.
