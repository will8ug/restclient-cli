## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init github.com/will8ug/restclient-cli`) and create directory structure: `cmd/`, `internal/parser/`, `internal/executor/`, `internal/tui/`
- [x] 1.2 Add dependencies: `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/lipgloss github.com/charmbracelet/bubbles`
- [x] 1.3 Scaffold `main.go` — parse CLI args (file path, `--version`), validate file exists, then launch `tea.NewProgram`
- [x] 1.4 Create `.gitignore` additions for Go (binary output) and a minimal `README.md`

## 2. HTTP File Parser

- [x] 2.1 Define data types: `Request` struct (Name, Method, URL, Headers, Body), `ParseResult` struct (Variables, Requests, Errors)
- [x] 2.2 Implement comment stripping — skip lines starting with `#` (but not `###`) or `//`
- [x] 2.3 Implement file-level variable parsing — extract `@name = value` definitions into a map
- [x] 2.4 Implement request splitting — split file content by `###` separator, extract optional request name from separator line
- [x] 2.5 Implement request line parsing — extract method and URL from the first non-blank, non-comment line of each block
- [x] 2.6 Implement header parsing — extract `Key: Value` lines after the request line, stop at blank line
- [x] 2.7 Implement body parsing — capture content after the first blank line until next separator or EOF
- [x] 2.8 Implement variable substitution — replace `{{name}}` in URLs, headers, and bodies; error on undefined variables
- [x] 2.9 Implement parse error reporting with line numbers
- [x] 2.10 Write unit tests for all parser requirements (multiple requests, comments, variables, headers, body, error cases)

## 3. Request Executor

- [x] 3.1 Implement `Execute(request Request, timeout time.Duration) (Response, error)` using `net/http` — send method, URL, headers, body
- [x] 3.2 Define `Response` struct (StatusCode, Status, Headers, Body, Duration)
- [x] 3.3 Implement timeout handling with configurable `time.Duration`
- [x] 3.4 Implement connection error handling with clear messages (DNS, connection refused)
- [x] 3.5 Configure redirect following (up to 10 hops) and redirect loop detection
- [x] 3.6 Implement response timing measurement (start-to-finish duration)
- [x] 3.7 Write unit tests for executor (use httptest for mock server — success, timeout, errors, redirects)

## 4. TUI — Bubble Tea App Model

- [x] 4.1 Define the root `Model` struct: parsed requests, selected index, active panel (enum: list/detail/response), loading state, current response, error state, window dimensions
- [x] 4.2 Implement `Init()` — return nil (no startup commands; requests are pre-parsed before TUI launch)
- [x] 4.3 Implement `Update()` — handle `tea.KeyMsg` for navigation (`↑`/`↓`/`j`/`k`), `Enter` for execute, `Tab` for panel switching, `q`/`Ctrl+C` for quit
- [x] 4.4 Implement `tea.Cmd` for async HTTP execution — fire request on `Enter`, return response as a custom `tea.Msg`
- [x] 4.5 Handle `tea.WindowSizeMsg` for responsive panel sizing on terminal resize

## 5. TUI — Panel Components

- [x] 5.1 Implement request list panel using `bubbles/list` — show index, name (or method+URL fallback), method badge with color coding (GET=green, POST=blue, PUT=yellow, DELETE=red)
- [x] 5.2 Implement request detail panel using `bubbles/viewport` — render selected request's method, URL, headers, and body as read-only scrollable content
- [x] 5.3 Implement response viewer panel using `bubbles/viewport` — render status line (color-coded: green 2xx, yellow 3xx, red 4xx/5xx), timing, headers, and body
- [x] 5.4 Implement JSON pretty-printing for response bodies with `application/json` content type
- [x] 5.5 Implement loading spinner using `bubbles/spinner` — show in response panel while request is in flight
- [x] 5.6 Implement error display in response panel — show connection errors, timeouts in red

## 6. TUI — Layout & Styling

- [x] 6.1 Implement three-panel `View()` layout using Lip Gloss — left panel (request list, ~30% width), right column split into top (request detail, ~40% height) and bottom (response viewer, ~60% height)
- [x] 6.2 Implement panel borders with focus indication — active panel gets a highlight color border, inactive panels get a dimmed border
- [x] 6.3 Implement status bar at the bottom — show key hints (`↑↓ navigate  enter send  tab switch  / filter  ? help  q quit`) and loaded file name
- [x] 6.4 Implement loading state in status bar — show spinner and request name when a request is in flight

## 7. TUI — Filtering & Help

- [x] 7.1 Implement request filtering — `/` activates filter input, filter request list by name/URL match, `Esc` clears filter
- [x] 7.2 Implement help overlay — `?` toggles a modal showing all keyboard shortcuts, `Esc` or `?` dismisses it

## 8. Integration & Polish

- [x] 8.1 Create sample `.http` files in `examples/` directory for manual testing
- [ ] 8.2 Write integration tests — parse file, verify model state transitions (select, execute, receive response)
- [x] 8.3 Verify `go build` produces a working binary and `go test ./...` passes
- [ ] 8.4 Test on multiple terminals (iTerm2, Terminal.app, Alacritty) for rendering compatibility
- [x] 8.5 Add build instructions, screenshots, and usage examples to `README.md`
