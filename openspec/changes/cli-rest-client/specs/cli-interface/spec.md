## ADDED Requirements

### Requirement: Run command with file argument
The CLI SHALL accept a `run` command with a file path argument to parse and execute requests from an `.http` file.

#### Scenario: Run all requests in file
- **WHEN** the user runs `restclient run api.http`
- **THEN** the tool parses `api.http` and executes all requests in sequence, displaying each response

#### Scenario: File not found
- **WHEN** the user runs `restclient run nonexistent.http`
- **THEN** the tool reports an error that the file was not found

#### Scenario: File with parse errors
- **WHEN** the `.http` file contains syntax errors
- **THEN** the tool reports the parse errors with line numbers and does not execute any requests

### Requirement: Select specific request by name or index
The CLI SHALL support a `--request` flag (short: `-r`) to run a specific request by its name (from `###` description) or 1-based index.

#### Scenario: Select by name
- **WHEN** the user runs `restclient run api.http -r "Get all users"`
- **THEN** only the request named "Get all users" is executed

#### Scenario: Select by index
- **WHEN** the user runs `restclient run api.http -r 2`
- **THEN** only the second request in the file is executed

#### Scenario: Request not found
- **WHEN** the user specifies a name or index that does not match any request
- **THEN** the tool reports an error listing available requests

### Requirement: Verbose output mode
The CLI SHALL support a `--verbose` flag (short: `-v`) that includes the full request details (method, URL, headers, body) in the output before the response.

#### Scenario: Verbose enabled
- **WHEN** the user runs with `--verbose`
- **THEN** the output shows the request method, URL, headers, and body, followed by the response

#### Scenario: Default (non-verbose)
- **WHEN** the user runs without `--verbose`
- **THEN** only the response is displayed

### Requirement: Custom timeout flag
The CLI SHALL support a `--timeout` flag (short: `-t`) to set the request timeout in seconds. Default is 30 seconds.

#### Scenario: Custom timeout
- **WHEN** the user runs `restclient run api.http --timeout 5`
- **THEN** each request uses a 5-second timeout

### Requirement: List requests in a file
The CLI SHALL support a `list` command that shows all requests in an `.http` file with their index and name.

#### Scenario: List requests
- **WHEN** the user runs `restclient list api.http`
- **THEN** the output shows each request's index, name (if any), method, and URL

### Requirement: Version flag
The CLI SHALL support a `--version` flag that prints the tool version and exits.

#### Scenario: Version output
- **WHEN** the user runs `restclient --version`
- **THEN** the tool prints its version string and exits

### Requirement: Help text
The CLI SHALL display help text describing all commands and flags when run with `--help` or with no arguments.

#### Scenario: Help flag
- **WHEN** the user runs `restclient --help`
- **THEN** comprehensive help text with all commands and flags is displayed

#### Scenario: No arguments
- **WHEN** the user runs `restclient` with no arguments
- **THEN** the help text is displayed

### Requirement: Exit codes
The CLI SHALL use exit code 0 for success and non-zero for failures.

#### Scenario: Successful execution
- **WHEN** all requests complete without errors
- **THEN** the exit code is 0

#### Scenario: Request failure
- **WHEN** any request fails (connection error, timeout, etc.)
- **THEN** the exit code is non-zero

#### Scenario: Parse error
- **WHEN** the `.http` file has parse errors
- **THEN** the exit code is non-zero
