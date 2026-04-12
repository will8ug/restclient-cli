package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	variablePattern = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.-]+)\s*\}\}`)
	validMethods    = map[string]struct{}{
		"GET":     {},
		"POST":    {},
		"PUT":     {},
		"PATCH":   {},
		"DELETE":  {},
		"HEAD":    {},
		"OPTIONS": {},
	}
)

// Request represents a parsed HTTP request from a .http file
type Request struct {
	Name    string            // from ### separator description (may be empty)
	Method  string            // GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
	URL     string            // fully resolved URL
	Headers map[string]string // request headers
	Body    string            // request body (may be empty)
}

// ParseError represents a parsing error with location
type ParseError struct {
	Line    int
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// ParseResult contains the result of parsing a .http file
type ParseResult struct {
	Requests []Request
	Errors   []ParseError
}

type rawLine struct {
	text string
	line int
}

type rawBlock struct {
	name  string
	lines []rawLine
}

type rawHeader struct {
	key   string
	value string
	line  int
}

// ParseFile reads and parses an .http file, returning all requests
func ParseFile(filename string) (*ParseResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(string(content)), nil
}

// Parse parses .http file content string
func Parse(content string) *ParseResult {
	result := &ParseResult{}
	variables := map[string]string{}
	blocks := splitBlocks(content, variables)

	for _, block := range blocks {
		request, errs := parseBlock(block, variables)
		if request != nil {
			result.Requests = append(result.Requests, *request)
		}
		result.Errors = append(result.Errors, errs...)
	}

	return result
}

func splitBlocks(content string, variables map[string]string) []rawBlock {
	lines := strings.Split(content, "\n")
	blocks := make([]rawBlock, 0)
	var current *rawBlock
	pendingName := ""

	flush := func() {
		if current == nil || len(current.lines) == 0 {
			current = nil
			return
		}
		blocks = append(blocks, *current)
		current = nil
	}

	for i, line := range lines {
		lineNumber := i + 1
		trimmed := strings.TrimSpace(line)

		if isSeparator(trimmed) {
			flush()
			pendingName = strings.TrimSpace(strings.TrimPrefix(trimmed, "###"))
			continue
		}

		if isComment(trimmed) {
			continue
		}

		if current == nil {
			if trimmed == "" {
				continue
			}

			if name, value, ok := parseVariableDefinition(trimmed); ok {
				variables[name] = value
				continue
			}

			current = &rawBlock{name: pendingName}
			pendingName = ""
		}

		current.lines = append(current.lines, rawLine{text: line, line: lineNumber})
	}

	flush()

	return blocks
}

func parseBlock(block rawBlock, variables map[string]string) (*Request, []ParseError) {
	requestLineIndex := -1
	for i, line := range block.lines {
		if strings.TrimSpace(line.text) != "" {
			requestLineIndex = i
			break
		}
	}

	if requestLineIndex == -1 {
		return nil, nil
	}

	requestLine := block.lines[requestLineIndex]
	fields := strings.Fields(requestLine.text)
	if len(fields) < 2 {
		return nil, []ParseError{{
			Line:    requestLine.line,
			Message: "missing request method",
		}}
	}

	method := strings.ToUpper(fields[0])
	if _, ok := validMethods[method]; !ok {
		return nil, []ParseError{{
			Line:    requestLine.line,
			Message: fmt.Sprintf("invalid method %q", fields[0]),
		}}
	}

	url, urlErrors := substituteVariables(fields[1], requestLine.line, variables)

	headers := map[string]string{}
	rawHeaders := make([]rawHeader, 0)
	bodyLines := make([]rawLine, 0)
	parsingBody := false

	for _, line := range block.lines[requestLineIndex+1:] {
		if parsingBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		if strings.TrimSpace(line.text) == "" {
			parsingBody = true
			continue
		}

		key, value, ok := parseHeader(line.text)
		if !ok {
			break
		}

		rawHeaders = append(rawHeaders, rawHeader{key: key, value: value, line: line.line})
	}

	parseErrors := append([]ParseError{}, urlErrors...)
	for _, header := range rawHeaders {
		resolvedKey, keyErrors := substituteVariables(header.key, header.line, variables)
		resolvedValue, valueErrors := substituteVariables(header.value, header.line, variables)
		parseErrors = append(parseErrors, keyErrors...)
		parseErrors = append(parseErrors, valueErrors...)
		headers[resolvedKey] = resolvedValue
	}

	resolvedBodyLines := make([]string, 0, len(bodyLines))
	for _, line := range bodyLines {
		resolvedLine, bodyErrors := substituteVariables(line.text, line.line, variables)
		parseErrors = append(parseErrors, bodyErrors...)
		resolvedBodyLines = append(resolvedBodyLines, resolvedLine)
	}

	request := &Request{
		Name:    block.name,
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    strings.TrimSpace(strings.Join(resolvedBodyLines, "\n")),
	}

	return request, parseErrors
}

func parseVariableDefinition(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "@") {
		return "", "", false
	}

	parts := strings.SplitN(line[1:], "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "", "", false
	}

	return name, strings.TrimSpace(parts[1]), true
}

func parseHeader(line string) (string, string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false
	}

	return key, strings.TrimSpace(parts[1]), true
}

func substituteVariables(value string, line int, variables map[string]string) (string, []ParseError) {
	errors := make([]ParseError, 0)
	resolved := variablePattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		name := parts[1]
		replacement, ok := variables[name]
		if !ok {
			errors = append(errors, ParseError{
				Line:    line,
				Message: fmt.Sprintf("undefined variable %q", name),
			})
			return match
		}

		return replacement
	})

	return resolved, errors
}

func isSeparator(line string) bool {
	return strings.HasPrefix(line, "###")
}

func isComment(line string) bool {
	if line == "" || isSeparator(line) {
		return false
	}

	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}
