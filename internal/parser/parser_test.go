package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSingleGetRequest(t *testing.T) {
	content := "GET https://api.example.com/users"
	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}

	req := result.Requests[0]
	if req.Method != "GET" {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("expected URL https://api.example.com/users, got %s", req.URL)
	}
	if req.Body != "" {
		t.Errorf("expected empty body, got %q", req.Body)
	}
}

func TestParseMultipleRequests(t *testing.T) {
	content := `GET https://api.example.com/users

###

POST https://api.example.com/users
Content-Type: application/json

{"name": "test"}

### Delete user

DELETE https://api.example.com/users/1`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(result.Requests))
	}

	if result.Requests[0].Method != "GET" {
		t.Errorf("req 0: expected GET, got %s", result.Requests[0].Method)
	}
	if result.Requests[1].Method != "POST" {
		t.Errorf("req 1: expected POST, got %s", result.Requests[1].Method)
	}
	if result.Requests[2].Method != "DELETE" {
		t.Errorf("req 2: expected DELETE, got %s", result.Requests[2].Method)
	}
	if result.Requests[2].Name != "Delete user" {
		t.Errorf("req 2: expected name 'Delete user', got %q", result.Requests[2].Name)
	}
}

func TestParseSeparatorWithName(t *testing.T) {
	content := `### Get all users
GET https://api.example.com/users`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}
	if result.Requests[0].Name != "Get all users" {
		t.Errorf("expected name 'Get all users', got %q", result.Requests[0].Name)
	}
}

func TestParseHeaders(t *testing.T) {
	content := `POST https://api.example.com/users
Content-Type: application/json
Authorization: Bearer token123

{"name": "test"}`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}

	req := result.Requests[0]
	if len(req.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(req.Headers))
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", req.Headers["Content-Type"])
	}
	if req.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected Authorization=Bearer token123, got %q", req.Headers["Authorization"])
	}
}

func TestParseNoHeaders(t *testing.T) {
	content := `GET https://api.example.com/users`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	req := result.Requests[0]
	if len(req.Headers) != 0 {
		t.Errorf("expected 0 headers, got %d", len(req.Headers))
	}
}

func TestParseBody(t *testing.T) {
	content := `POST https://api.example.com/users
Content-Type: application/json

{
  "name": "test",
  "email": "test@example.com"
}`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	expected := `{
  "name": "test",
  "email": "test@example.com"
}`
	if result.Requests[0].Body != expected {
		t.Errorf("expected body:\n%s\ngot:\n%s", expected, result.Requests[0].Body)
	}
}

func TestParseNoBody(t *testing.T) {
	content := `GET https://api.example.com/users
Accept: application/json`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Requests[0].Body != "" {
		t.Errorf("expected empty body, got %q", result.Requests[0].Body)
	}
}

func TestParseComments(t *testing.T) {
	content := `# This is a comment
// This is also a comment
GET https://api.example.com/users
# Another comment`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}
	if result.Requests[0].Method != "GET" {
		t.Errorf("expected GET, got %s", result.Requests[0].Method)
	}
}

func TestParseSeparatorNotComment(t *testing.T) {
	content := `### First request
GET https://api.example.com/first

### Second request
GET https://api.example.com/second`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(result.Requests))
	}
}

func TestParseVariables(t *testing.T) {
	content := `@host = https://api.example.com
@token = abc123

### Get users
GET {{host}}/users
Authorization: Bearer {{token}}`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}

	req := result.Requests[0]
	if req.URL != "https://api.example.com/users" {
		t.Errorf("expected resolved URL, got %s", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer abc123" {
		t.Errorf("expected resolved header, got %q", req.Headers["Authorization"])
	}
}

func TestParseVariableInBody(t *testing.T) {
	content := `@userId = 42

POST https://api.example.com/users
Content-Type: application/json

{"id": "{{userId}}"}`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	if result.Requests[0].Body != `{"id": "42"}` {
		t.Errorf("expected resolved body, got %q", result.Requests[0].Body)
	}
}

func TestParseUndefinedVariable(t *testing.T) {
	content := `GET https://{{undefined_host}}/users`

	result := Parse(content)

	if len(result.Errors) == 0 {
		t.Fatal("expected error for undefined variable")
	}

	found := false
	for _, e := range result.Errors {
		if contains(e.Message, "undefined") && contains(e.Message, "undefined_host") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about undefined_host, got: %v", result.Errors)
	}
}

func TestParseInvalidMethod(t *testing.T) {
	content := `INVALID https://api.example.com/users`

	result := Parse(content)

	if len(result.Errors) == 0 {
		t.Fatal("expected error for invalid method")
	}

	found := false
	for _, e := range result.Errors {
		if contains(e.Message, "invalid") || contains(e.Message, "INVALID") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about invalid method, got: %v", result.Errors)
	}
}

func TestParseErrorLineNumber(t *testing.T) {
	content := `
INVALID https://api.example.com/users`

	result := Parse(content)

	if len(result.Errors) == 0 {
		t.Fatal("expected error")
	}
	if result.Errors[0].Line < 1 {
		t.Errorf("expected valid line number, got %d", result.Errors[0].Line)
	}
}

func TestParseEmptyFile(t *testing.T) {
	result := Parse("")

	if len(result.Errors) != 0 {
		t.Errorf("expected no errors for empty file, got: %v", result.Errors)
	}
	if len(result.Requests) != 0 {
		t.Errorf("expected no requests for empty file, got %d", len(result.Requests))
	}
}

func TestParseAllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		content := method + " https://api.example.com/test"
		result := Parse(content)

		if len(result.Errors) > 0 {
			t.Errorf("method %s: unexpected errors: %v", method, result.Errors)
		}
		if len(result.Requests) != 1 {
			t.Errorf("method %s: expected 1 request, got %d", method, len(result.Requests))
			continue
		}
		if result.Requests[0].Method != method {
			t.Errorf("expected method %s, got %s", method, result.Requests[0].Method)
		}
	}
}

func TestParseRequestWithHTTPVersion(t *testing.T) {
	content := `POST https://api.example.com/users HTTP/1.1
Content-Type: application/json

{"name": "test"}`

	result := Parse(content)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}
	if result.Requests[0].URL != "https://api.example.com/users" {
		t.Errorf("expected URL without HTTP version, got %s", result.Requests[0].URL)
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.http")
	content := `GET https://api.example.com/users`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if len(result.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Requests))
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/test.http")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
