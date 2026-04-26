# restclient-cli

A TUI (Terminal User Interface) REST API client that reads `.http` files — the same format used by VS Code's REST Client extension.

## Features

- Parse `.http` files with requests, headers, bodies, comments, and variables
- Interactive three-panel TUI: request list, request detail, response viewer
- Keyboard-driven navigation
- Colored method badges and status codes
- JSON pretty-printing for response bodies
- Request filtering by name/URL
- Variable substitution (`@name = value` / `{{name}}`)

## Build

```bash
# With Make
make build          # → ./bin/restclient-cli
make install        # → $GOPATH/bin/restclient-cli

# Or directly
go build -o restclient-cli ./cmd/
```

## Usage

```bash
# Launch TUI with an .http file
./restclient-cli examples/jsonplaceholder.http

# Show version
./restclient-cli --version
```

## Development

```bash
make test          # go test -race -count=1 ./...
make vet           # go vet ./...
make fmt           # go fmt ./...
make lint          # golangci-lint run (requires install)
make clean         # remove ./bin/
```

> **Prerequisites**: Go 1.26+, [golangci-lint](https://golangci-lint.run/) (optional, for `make lint`)

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate / scroll |
| `Enter` | Send selected request |
| `Tab` | Switch panel focus |
| `/` | Filter requests |
| `?` | Toggle help |
| `q` | Quit |
| `Ctrl+C` | Force quit |

## .http File Format

```http
@host = https://api.example.com
@token = my-api-token

### Get all users
GET {{host}}/users
Authorization: Bearer {{token}}

### Create user
POST {{host}}/users
Content-Type: application/json

{
  "name": "John",
  "email": "john@example.com"
}
```

## License

MIT
