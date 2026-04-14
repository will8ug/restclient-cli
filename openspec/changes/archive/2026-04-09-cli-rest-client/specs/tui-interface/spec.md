## ADDED Requirements

### Requirement: Launch TUI with file argument
The application SHALL accept a file path as a command-line argument, parse the `.http` file, and launch the interactive TUI displaying all parsed requests.

#### Scenario: Launch with valid file
- **WHEN** the user runs `restclient api.http`
- **THEN** the TUI launches showing the parsed requests in the request list panel

#### Scenario: File not found
- **WHEN** the user runs `restclient nonexistent.http`
- **THEN** the application prints an error to stderr and exits with a non-zero exit code (no TUI launched)

#### Scenario: File with parse errors
- **WHEN** the `.http` file contains syntax errors
- **THEN** the application prints the parse errors with line numbers to stderr and exits with a non-zero exit code

#### Scenario: No arguments
- **WHEN** the user runs `restclient` with no arguments
- **THEN** the application prints usage help and exits

### Requirement: Three-panel layout
The TUI SHALL display a three-panel layout: a request list on the left, a request detail panel on the top-right, and a response viewer on the bottom-right.

#### Scenario: Initial layout
- **WHEN** the TUI launches with a parsed `.http` file
- **THEN** the left panel shows the list of requests (index, name, method, URL), the top-right panel shows the first request's details, and the bottom-right panel is empty with a placeholder message

#### Scenario: Terminal resize
- **WHEN** the terminal window is resized
- **THEN** all panels adjust proportionally to fit the new dimensions

### Requirement: Request list navigation
The TUI SHALL allow navigating the request list using `↑`/`↓` arrow keys or `j`/`k` keys. The currently selected request is visually highlighted.

#### Scenario: Navigate down
- **WHEN** the user presses `↓` or `j`
- **THEN** the selection moves to the next request and the detail panel updates to show that request

#### Scenario: Navigate up
- **WHEN** the user presses `↑` or `k`
- **THEN** the selection moves to the previous request and the detail panel updates

#### Scenario: Boundary at top
- **WHEN** the first request is selected and the user presses `↑`
- **THEN** the selection remains on the first request (no wrap)

#### Scenario: Boundary at bottom
- **WHEN** the last request is selected and the user presses `↓`
- **THEN** the selection remains on the last request (no wrap)

### Requirement: Request detail panel
The TUI SHALL display the selected request's full details in the top-right panel: method, URL, headers, and body.

#### Scenario: Request with headers and body
- **WHEN** a request with headers and a JSON body is selected
- **THEN** the detail panel shows the method and URL on the first line, followed by headers, a blank line, and the body

#### Scenario: Request without body
- **WHEN** a GET request with no body is selected
- **THEN** the detail panel shows the method, URL, and headers only

### Requirement: Execute request on Enter
The TUI SHALL execute the currently selected request when the user presses `Enter`. While the request is in flight, a loading indicator is shown in the response panel.

#### Scenario: Execute and display response
- **WHEN** the user presses `Enter` on a selected request
- **THEN** a loading spinner appears in the response panel, and when the response arrives, the panel shows the status code, headers, and body

#### Scenario: Execute during loading
- **WHEN** a request is already in flight and the user presses `Enter`
- **THEN** the new request replaces the in-flight one (or the keypress is ignored until completion)

### Requirement: Response viewer panel
The TUI SHALL display the HTTP response in the bottom-right panel with the status line, response headers, and body. The panel is scrollable for long responses.

#### Scenario: JSON response formatting
- **WHEN** the response has `Content-Type: application/json`
- **THEN** the body is pretty-printed with indentation

#### Scenario: Scrollable response
- **WHEN** the response body exceeds the panel height
- **THEN** the user can scroll the response panel using arrow keys or `j`/`k` when the panel is focused

#### Scenario: Status code coloring
- **WHEN** the response status code is 2xx
- **THEN** the status line is displayed in green
- **WHEN** the response status code is 3xx
- **THEN** the status line is displayed in yellow
- **WHEN** the response status code is 4xx or 5xx
- **THEN** the status line is displayed in red

#### Scenario: Response timing
- **WHEN** a response is displayed
- **THEN** the elapsed time in milliseconds is shown next to the status code

#### Scenario: Error response
- **WHEN** the request fails (timeout, DNS error, connection refused)
- **THEN** the response panel shows the error message in red

### Requirement: Panel focus switching
The TUI SHALL support switching focus between panels using the `Tab` key. The focused panel is visually indicated (e.g., highlighted border).

#### Scenario: Tab cycling
- **WHEN** the user presses `Tab`
- **THEN** focus moves from request list → request detail → response viewer → request list (cyclic)

#### Scenario: Focused panel indication
- **WHEN** a panel receives focus
- **THEN** its border changes to a highlight color to indicate active focus

#### Scenario: Key behavior per panel
- **WHEN** the request list is focused
- **THEN** `↑`/`↓`/`j`/`k` navigate requests and `Enter` executes
- **WHEN** the request detail or response viewer is focused
- **THEN** `↑`/`↓`/`j`/`k` scroll the panel content

### Requirement: Request list filtering
The TUI SHALL support filtering the request list by pressing `/` and typing a search query. Only requests whose name or URL matches the query are shown.

#### Scenario: Filter by name
- **WHEN** the user presses `/` and types `user`
- **THEN** only requests with "user" in their name or URL are displayed

#### Scenario: Clear filter
- **WHEN** the user presses `Esc` while the filter input is active
- **THEN** the filter is cleared and all requests are shown again

### Requirement: Help overlay
The TUI SHALL display a help overlay listing all keyboard shortcuts when the user presses `?`.

#### Scenario: Show help
- **WHEN** the user presses `?`
- **THEN** a modal overlay appears showing all keyboard shortcuts

#### Scenario: Dismiss help
- **WHEN** the help overlay is visible and the user presses `?` or `Esc`
- **THEN** the overlay is dismissed and the TUI returns to normal

### Requirement: Status bar
The TUI SHALL display a status bar at the bottom showing contextual keyboard hints and the loaded file name.

#### Scenario: Default status bar
- **WHEN** the TUI is in normal mode
- **THEN** the status bar shows key hints (e.g., `↑↓ navigate  enter send  tab switch  / filter  ? help  q quit`) and the file name

#### Scenario: Loading state
- **WHEN** a request is in flight
- **THEN** the status bar shows a loading indicator with the request being executed

### Requirement: Quit application
The TUI SHALL exit cleanly when the user presses `q` or `Ctrl+C`.

#### Scenario: Quit with q
- **WHEN** the user presses `q` (while not in filter mode)
- **THEN** the TUI exits and the terminal is restored to its previous state

#### Scenario: Quit with Ctrl+C
- **WHEN** the user presses `Ctrl+C`
- **THEN** the TUI exits and the terminal is restored

### Requirement: Version flag
The application SHALL support a `--version` flag that prints the tool version and exits without launching the TUI.

#### Scenario: Version output
- **WHEN** the user runs `restclient --version`
- **THEN** the tool prints its version string and exits

### Requirement: Exit codes
The application SHALL use exit code 0 for success and non-zero for failures.

#### Scenario: Successful execution
- **WHEN** the TUI launches and the user quits normally
- **THEN** the exit code is 0

#### Scenario: Startup failure
- **WHEN** the file is not found or has parse errors
- **THEN** the exit code is non-zero
