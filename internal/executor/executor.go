package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

const defaultTimeout = 30 * time.Second

var transportFactory = func() http.RoundTripper {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultTransport
	}

	return transport.Clone()
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	Duration   time.Duration
}

// Request represents the input for execution.
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// Execute performs an HTTP request and returns the response.
func Execute(req Request, timeout time.Duration) (*Response, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	body := strings.NewReader(req.Body)
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transportFactory(),
		CheckRedirect: func(redirReq *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("redirect loop detected after 10 redirects")
			}
			return nil
		},
	}

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, wrapExecuteError(req.URL, timeout, err)
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(httpResp.Body)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header.Clone(),
		Body:       string(responseBody),
		Duration:   duration,
	}, nil
}

func wrapExecuteError(rawURL string, timeout time.Duration, err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return fmt.Errorf("execute request: %w", err)
	}

	target, parseErr := url.Parse(rawURL)
	host := ""
	if parseErr == nil {
		host = target.Host
	}

	if urlErr.Timeout() || errors.Is(urlErr.Err, context.DeadlineExceeded) {
		return fmt.Errorf("request timed out after %s: %w", timeout, err)
	}

	var dnsErr *net.DNSError
	if errors.As(urlErr.Err, &dnsErr) {
		name := dnsErr.Name
		if name == "" {
			name = target.Hostname()
		}
		return fmt.Errorf("dns resolution failed for %s: %w", name, err)
	}

	if errors.Is(urlErr.Err, syscall.ECONNREFUSED) || strings.Contains(strings.ToLower(urlErr.Err.Error()), "connection refused") {
		if host == "" {
			host = rawURL
		}
		return fmt.Errorf("connection refused for %s: %w", host, err)
	}

	var opErr *net.OpError
	if errors.As(urlErr.Err, &opErr) && errors.Is(opErr.Err, syscall.ECONNREFUSED) {
		if host == "" {
			host = rawURL
		}
		return fmt.Errorf("connection refused for %s: %w", host, err)
	}

	return fmt.Errorf("execute request: %w", err)
}
