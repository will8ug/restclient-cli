## ADDED Requirements

### Requirement: Execute HTTP request
The executor SHALL take a parsed request (method, URL, headers, body) and perform the corresponding HTTP request, returning the response status code, headers, and body.

#### Scenario: Successful GET request
- **WHEN** executing a GET request to a reachable URL
- **THEN** the executor returns the HTTP status code, response headers, and response body

#### Scenario: POST with body
- **WHEN** executing a POST request with a JSON body and Content-Type header
- **THEN** the executor sends the body and Content-Type header, and returns the response

#### Scenario: All supported methods
- **WHEN** executing a request with any of GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- **THEN** the executor sends the request using the specified method

### Requirement: Handle request timeouts
The executor SHALL enforce a configurable timeout. The default timeout is 30 seconds.

#### Scenario: Request exceeds timeout
- **WHEN** a request does not receive a response within the timeout period
- **THEN** the executor returns a timeout error with the duration that elapsed

#### Scenario: Custom timeout
- **WHEN** the user specifies a timeout of 5 seconds and the request takes 10 seconds
- **THEN** the executor returns a timeout error after 5 seconds

### Requirement: Handle connection errors
The executor SHALL report clear error messages for common failure modes.

#### Scenario: DNS resolution failure
- **WHEN** the request URL contains an unresolvable hostname
- **THEN** the executor reports a DNS resolution error with the hostname

#### Scenario: Connection refused
- **WHEN** the target host refuses the connection
- **THEN** the executor reports a connection refused error with host and port

### Requirement: Follow redirects
The executor SHALL follow HTTP redirects (3xx) by default, up to a maximum of 10 redirects.

#### Scenario: 301 redirect
- **WHEN** the server responds with 301 and a Location header
- **THEN** the executor follows the redirect and returns the final response

#### Scenario: Redirect loop
- **WHEN** redirects exceed 10 hops
- **THEN** the executor returns an error indicating too many redirects

### Requirement: Support HTTPS
The executor SHALL support HTTPS URLs using the system's default TLS configuration.

#### Scenario: HTTPS request
- **WHEN** the request URL uses the `https://` scheme
- **THEN** the executor performs a TLS handshake and completes the request

### Requirement: Report response timing
The executor SHALL measure and report the total time taken for each request (from connection start to response body fully received).

#### Scenario: Timing reported
- **WHEN** a request completes successfully
- **THEN** the response includes the elapsed time in milliseconds
