package executor

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExecuteGetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("X-Test", "value")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	resp, err := Execute(Request{Method: "GET", URL: server.URL + "/test"}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Body != `{"status":"ok"}` {
		t.Errorf("expected body, got %q", resp.Body)
	}
	if resp.Headers.Get("X-Test") != "value" {
		t.Errorf("expected X-Test header, got %q", resp.Headers.Get("X-Test"))
	}
}

func TestExecutePostWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	resp, err := Execute(Request{
		Method:  "POST",
		URL:     server.URL + "/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"test"}`,
	}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestExecuteAllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, method := range methods {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				t.Errorf("expected %s, got %s", method, r.Method)
			}
			w.WriteHeader(200)
		}))

		_, err := Execute(Request{Method: method, URL: server.URL}, 5*time.Second)
		server.Close()
		if err != nil {
			t.Errorf("method %s: unexpected error: %v", method, err)
		}
	}
}

func TestExecuteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(200)
	}))
	defer server.Close()

	_, err := Execute(Request{Method: "GET", URL: server.URL}, 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %v", err)
	}
}

func TestExecuteConnectionRefused(t *testing.T) {
	// Use a port that's not listening
	_, err := Execute(Request{Method: "GET", URL: "http://127.0.0.1:1"}, 2*time.Second)
	if err == nil {
		t.Fatal("expected connection error")
	}
	if !strings.Contains(err.Error(), "connection refused") && !strings.Contains(err.Error(), "execute request") {
		t.Errorf("expected connection refused message, got: %v", err)
	}
}

func TestExecuteDNSFailure(t *testing.T) {
	_, err := Execute(Request{Method: "GET", URL: "http://this-host-does-not-exist-xyz.invalid/test"}, 5*time.Second)
	if err == nil {
		t.Fatal("expected DNS error")
	}
	if !strings.Contains(err.Error(), "dns") && !strings.Contains(err.Error(), "no such host") && !strings.Contains(err.Error(), "execute request") {
		t.Errorf("expected DNS error message, got: %v", err)
	}
}

func TestExecuteRedirect(t *testing.T) {
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("final"))
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	resp, err := Execute(Request{Method: "GET", URL: redirectServer.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after redirect, got %d", resp.StatusCode)
	}
	if resp.Body != "final" {
		t.Errorf("expected 'final' body, got %q", resp.Body)
	}
}

func TestExecuteRedirectLoop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	}))
	defer server.Close()

	_, err := Execute(Request{Method: "GET", URL: server.URL}, 5*time.Second)
	if err == nil {
		t.Fatal("expected redirect loop error")
	}
}

func TestExecuteResponseTiming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	resp, err := Execute(Request{Method: "GET", URL: server.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", resp.Duration)
	}
	if resp.Duration < 50*time.Millisecond {
		t.Errorf("expected duration >= 50ms, got %v", resp.Duration)
	}
}

func TestExecuteDefaultTimeout(t *testing.T) {
	// Passing 0 should use default timeout (30s), not hang forever
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	resp, err := Execute(Request{Method: "GET", URL: server.URL}, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestExecuteHTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("secure"))
	}))
	defer server.Close()

	// Override transport to trust the test server's TLS cert
	origFactory := transportFactory
	transportFactory = func() http.RoundTripper {
		return &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	defer func() { transportFactory = origFactory }()

	resp, err := Execute(Request{Method: "GET", URL: server.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Body != "secure" {
		t.Errorf("expected 'secure', got %q", resp.Body)
	}
}
