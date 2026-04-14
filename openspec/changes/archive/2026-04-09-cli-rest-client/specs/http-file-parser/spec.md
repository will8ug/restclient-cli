## ADDED Requirements

### Requirement: Parse request separator
The parser SHALL split an `.http` file into individual requests using `###` as the delimiter. Each `###` line MAY include a trailing description that serves as the request name.

#### Scenario: File with multiple requests
- **WHEN** an `.http` file contains two requests separated by `###`
- **THEN** the parser produces two distinct request objects

#### Scenario: First request without leading separator
- **WHEN** an `.http` file starts with a request line (no `###` before it)
- **THEN** the parser treats it as the first request

#### Scenario: Separator with description
- **WHEN** a separator line is `### Get all users`
- **THEN** the request following it has the name `Get all users`

### Requirement: Parse request line
The parser SHALL extract the HTTP method and URL from the first non-comment, non-variable, non-blank line of each request block. The format is `METHOD URL` or `METHOD URL HTTP/version`. The HTTP version is optional and ignored.

#### Scenario: Simple GET request
- **WHEN** the request line is `GET https://api.example.com/users`
- **THEN** the parser produces a request with method `GET` and URL `https://api.example.com/users`

#### Scenario: Request with HTTP version
- **WHEN** the request line is `POST https://api.example.com/users HTTP/1.1`
- **THEN** the parser produces a request with method `POST` and URL `https://api.example.com/users`

#### Scenario: Supported HTTP methods
- **WHEN** the request line uses any of GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- **THEN** the parser accepts it as a valid method

### Requirement: Parse headers
The parser SHALL extract headers from lines following the request line that match the pattern `Name: Value`. Header parsing stops at the first blank line or at the next request separator.

#### Scenario: Single header
- **WHEN** a line after the request line is `Content-Type: application/json`
- **THEN** the parser adds header `Content-Type` with value `application/json` to the request

#### Scenario: Multiple headers
- **WHEN** three header lines follow the request line
- **THEN** all three headers are captured in the request

#### Scenario: No headers
- **WHEN** a blank line immediately follows the request line
- **THEN** the request has no custom headers

### Requirement: Parse request body
The parser SHALL treat all content after the first blank line (following headers) as the request body, up to the next `###` separator or end of file. Leading and trailing whitespace of the body block is trimmed.

#### Scenario: JSON body
- **WHEN** the content after the blank line is `{"name": "test"}`
- **THEN** the request body is `{"name": "test"}`

#### Scenario: No body
- **WHEN** there is no blank line after the headers (or the blank line is immediately followed by `###`)
- **THEN** the request body is empty

#### Scenario: Multi-line body
- **WHEN** the body spans multiple lines
- **THEN** the parser preserves all lines as the body content with their original line breaks

### Requirement: Parse comments
The parser SHALL ignore lines that start with `#` (but not `###`) or `//`. Comments can appear anywhere in the file.

#### Scenario: Hash comment
- **WHEN** a line is `# This is a comment`
- **THEN** the parser skips this line

#### Scenario: Double-slash comment
- **WHEN** a line is `// Another comment`
- **THEN** the parser skips this line

#### Scenario: Separator is not a comment
- **WHEN** a line is `###` or `### Name`
- **THEN** the parser treats it as a request separator, NOT a comment

### Requirement: Parse and substitute variables
The parser SHALL extract file-level variable definitions in the format `@name = value` and substitute all occurrences of `{{name}}` in URLs, headers, and bodies with the corresponding value.

#### Scenario: Variable definition and usage
- **WHEN** the file contains `@host = https://api.example.com` and a request URL is `{{host}}/users`
- **THEN** the resolved URL is `https://api.example.com/users`

#### Scenario: Variable in header value
- **WHEN** a variable `@token = abc123` is defined and a header is `Authorization: Bearer {{token}}`
- **THEN** the resolved header value is `Bearer abc123`

#### Scenario: Variable in body
- **WHEN** a variable `@userId = 42` is defined and the body contains `{"id": "{{userId}}"}`
- **THEN** the resolved body is `{"id": "42"}`

#### Scenario: Undefined variable
- **WHEN** the file references `{{undefined_var}}` but no matching `@undefined_var` definition exists
- **THEN** the parser reports an error identifying the undefined variable name

### Requirement: Report parse errors with location
The parser SHALL report errors with the line number where the problem occurred.

#### Scenario: Missing request method
- **WHEN** a request block has no valid request line
- **THEN** the parser reports an error with the line number

#### Scenario: Invalid method
- **WHEN** the request line uses an unrecognized HTTP method
- **THEN** the parser reports an error naming the invalid method and its line number
