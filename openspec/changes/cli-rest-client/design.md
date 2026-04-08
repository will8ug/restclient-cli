## Context

This is a greenfield project. No existing codebase, no legacy constraints. The goal is a CLI tool that understands the `.http` file format from VS Code's REST Client extension and executes requests from the terminal.

The `.http` format is a de facto standard across VS Code REST Client and JetBrains HTTP Client. It defines HTTP requests in plain text files with a simple, human-readable syntax. Currently, executing these files requires an IDE — there's no good standalone CLI option.

## Goals / Non-Goals

**Goals:**
- Parse `.http` files compatible with VS Code REST Client's core syntax
- Execute HTTP requests and display formatted responses
- Support file-level variables and substitution
- Run as a single binary from the command line with no runtime dependencies
- Keep the initial scope minimal — cover the 80% use case

**Non-Goals:**
- Environment files (`.env`) or multi-environment switching (future)
- OAuth / authentication helpers (future)
- Request history or cookie jar persistence (future)
- Response scripting or assertions (future)
- GraphQL-specific support (future)
- Certificate / mTLS configuration (future)
- Request chaining (using response values in subsequent requests) (future)

## Decisions

### 1. Language: Go

**Choice**: Go

**Rationale**: Go produces a single static binary with no runtime dependencies, has excellent standard library HTTP support (`net/http`), fast compilation, and cross-platform builds are trivial (`GOOS`/`GOARCH`). The CLI ecosystem is mature (cobra, etc.).

**Alternatives considered**:
- **Rust**: Better performance but slower development iteration; overkill for an HTTP client tool
- **TypeScript/Node**: Requires Node.js runtime installed; defeats the "single binary" goal
- **Python**: Requires Python runtime; distribution is messy

### 2. CLI Framework: Cobra

**Choice**: `github.com/spf13/cobra`

**Rationale**: De facto standard for Go CLIs. Provides subcommand routing, flag parsing, help generation, and shell completions out of the box.

### 3. HTTP Client: net/http (standard library)

**Choice**: Go's built-in `net/http`

**Rationale**: Covers all HTTP methods, custom headers, request bodies, and timeouts. No need for a third-party HTTP library for basic functionality.

### 4. Architecture: Three-layer pipeline

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Parser     │ →  │   Executor   │ →  │   Formatter  │
│ (.http → AST)│    │ (AST → HTTP) │    │ (Response →  │
│              │    │              │    │   stdout)    │
└──────────────┘    └──────────────┘    └──────────────┘
```

- **Parser**: Reads `.http` file, produces a list of `Request` structs with resolved variables
- **Executor**: Takes a `Request`, makes the HTTP call, returns a `Response`
- **Formatter**: Takes a `Response`, formats it for terminal output

Each layer is a clean interface — easily testable and replaceable.

### 5. .http File Format Support (initial scope)

Supported syntax:
- Request separator: `###` (with optional description after)
- Request line: `METHOD URL HTTP/version` (HTTP version optional)
- Headers: `Header-Name: value` (one per line, immediately after request line)
- Body: everything after the first blank line until next `###` or EOF
- Comments: lines starting with `#` or `//`
- Variables: `@name = value` at file level, referenced as `{{name}}`

### 6. Output Format

Default: colored terminal output showing status line, response headers, and body. JSON bodies are pretty-printed. A `--verbose` flag shows request details too.

## Risks / Trade-offs

- **Format compatibility**: The `.http` format isn't formally specified. Edge cases in VS Code REST Client's parser may differ from ours. → **Mitigation**: Start with the well-documented core syntax; add compatibility tests against real-world `.http` files over time.
- **Single binary size**: Go binaries are larger than C/Rust (~10-15MB). → **Acceptable**: Still a single file, trivial to distribute.
- **No streaming**: Initial implementation reads entire response into memory. → **Mitigation**: Fine for API testing; add streaming support later if needed for large responses.
