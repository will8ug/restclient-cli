## Why

There is no lightweight, standalone terminal tool that natively understands the `.http` file format popularized by VS Code's REST Client extension. Developers who write `.http` files for API testing are locked into VS Code or JetBrains IDEs to execute them. A TUI (Terminal User Interface) tool would provide an interactive, visual experience for browsing, editing, and executing requests — all without leaving the terminal. This enables faster API workflows for terminal-native developers.

## What Changes

- Introduce a new TUI tool (`restclient-cli`) that parses `.http` files and presents them in an interactive terminal interface
- Support the core `.http` file format: HTTP method + URL, headers, request body, and request separators (`###`)
- Support file-level variables (`@name = value`) and variable substitution (`{{name}}`)
- Support comments (`#` and `//`)
- Provide a multi-panel TUI layout: request list, request detail, and response viewer
- Display formatted, syntax-highlighted responses with status codes, headers, and pretty-printed bodies
- Keyboard-driven navigation for selecting and executing requests

## Capabilities

### New Capabilities
- `http-file-parser`: Parse `.http` file format — requests, headers, body, separators (`###`), comments, and variables
- `request-executor`: Execute HTTP requests (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) and return structured responses
- `tui-interface`: Interactive terminal user interface with multi-panel layout, keyboard navigation, and formatted response display

### Modified Capabilities

(none — this is a greenfield project)

## Impact

- **New project**: Entire codebase is new — no existing code affected
- **Dependencies**: Bubble Tea (TUI framework), Lip Gloss (styling), Bubbles (components), net/http (HTTP client)
- **APIs**: No external APIs consumed beyond the target URLs in `.http` files
- **Systems**: Runs locally as a standalone TUI binary
