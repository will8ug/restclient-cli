## Why

There is no lightweight, standalone CLI tool that natively understands the `.http` file format popularized by VS Code's REST Client extension. Developers who write `.http` files for API testing are locked into VS Code or JetBrains IDEs to execute them. A pure CLI tool would enable running these files in CI/CD pipelines, terminal workflows, and editor-agnostic environments.

## What Changes

- Introduce a new CLI tool (`restclient-cli`) that parses and executes `.http` files from the command line
- Support the core `.http` file format: HTTP method + URL, headers, request body, and request separators (`###`)
- Support file-level variables (`@name = value`) and variable substitution (`{{name}}`)
- Support comments (`#` and `//`)
- Provide clear, formatted output of responses (status, headers, body)
- Support selecting and running specific requests from a multi-request `.http` file

## Capabilities

### New Capabilities
- `http-file-parser`: Parse `.http` file format — requests, headers, body, separators (`###`), comments, and variables
- `request-executor`: Execute HTTP requests (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) and display formatted responses
- `cli-interface`: Command-line interface for selecting files, choosing requests, and controlling output

### Modified Capabilities

(none — this is a greenfield project)

## Impact

- **New project**: Entire codebase is new — no existing code affected
- **Dependencies**: Will need an HTTP client library and a CLI argument parser (language TBD in design)
- **APIs**: No external APIs consumed beyond the target URLs in `.http` files
- **Systems**: Runs locally as a standalone CLI binary
