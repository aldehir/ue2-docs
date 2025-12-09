package fetcher

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aldehir/ue2-docs/internal/urlutil"
)

func TestFetcher_Fetch_Success(t *testing.T) {
	expectedBody := []byte("test response")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		if ua := r.Header.Get("User-Agent"); ua != "test-agent" {
			t.Errorf("expected User-Agent 'test-agent', got '%s'", ua)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedBody)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.UserAgent = "test-agent"
	fetcher := New(config)

	ctx := context.Background()
	resp, err := fetcher.Fetch(ctx, server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if string(resp.Body) != string(expectedBody) {
		t.Errorf("expected body '%s', got '%s'", expectedBody, resp.Body)
	}

	if resp.ContentType != "text/html; charset=utf-8" {
		t.Errorf("expected content-type 'text/html; charset=utf-8', got '%s'", resp.ContentType)
	}

	if resp.ResourceType != urlutil.ResourceHTML {
		t.Errorf("expected resource type HTML, got %v", resp.ResourceType)
	}
}

func TestFetcher_Fetch_RetryOnServerError(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)

		// Fail first 2 attempts, succeed on 3rd
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxRetries = 3
	config.InitialDelay = 10 * time.Millisecond
	fetcher := New(config)

	ctx := context.Background()
	resp, err := fetcher.Fetch(ctx, server.URL)

	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestFetcher_Fetch_NoRetryOnClientError(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxRetries = 3
	fetcher := New(config)

	ctx := context.Background()
	_, err := fetcher.Fetch(ctx, server.URL)

	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}

	// Should only attempt once, no retries on 4xx errors
	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt (no retries on 404), got %d", attempts.Load())
	}
}

func TestFetcher_Fetch_MaxRetriesExceeded(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxRetries = 2
	config.InitialDelay = 10 * time.Millisecond
	fetcher := New(config)

	ctx := context.Background()
	_, err := fetcher.Fetch(ctx, server.URL)

	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}

	// Should attempt initial + 2 retries = 3 total
	expected := int32(3)
	if attempts.Load() != expected {
		t.Errorf("expected %d attempts, got %d", expected, attempts.Load())
	}
}

func TestFetcher_Fetch_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	fetcher := New(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := fetcher.Fetch(ctx, server.URL)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestFetcher_Fetch_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than timeout
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRetries = 0 // No retries
	fetcher := New(config)

	ctx := context.Background()
	_, err := fetcher.Fetch(ctx, server.URL)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestFetcher_Fetch_WithRateLimiter(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Allow 5 requests per 100ms
	rateLimiter := NewSimpleRateLimiter(5, 100*time.Millisecond)
	defer rateLimiter.Stop()

	config := DefaultConfig()
	config.RateLimiter = rateLimiter
	fetcher := New(config)

	ctx := context.Background()
	start := time.Now()

	// Make 10 requests
	for i := 0; i < 10; i++ {
		_, err := fetcher.Fetch(ctx, server.URL)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	elapsed := time.Since(start)

	// Should take at least 100ms due to rate limiting (5 initial + wait for 5 more)
	if elapsed < 80*time.Millisecond {
		t.Errorf("requests completed too quickly: %v (expected at least 80ms)", elapsed)
	}

	if requests.Load() != 10 {
		t.Errorf("expected 10 requests, got %d", requests.Load())
	}
}

func TestFetcher_CalculateBackoff(t *testing.T) {
	config := DefaultConfig()
	config.InitialDelay = 1 * time.Second
	config.MaxDelay = 30 * time.Second
	fetcher := New(config)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 1 * time.Second},   // 2^0 = 1
		{2, 2 * time.Second},   // 2^1 = 2
		{3, 4 * time.Second},   // 2^2 = 4
		{4, 8 * time.Second},   // 2^3 = 8
		{5, 16 * time.Second},  // 2^4 = 16
		{6, 30 * time.Second},  // 2^5 = 32, capped at 30
		{10, 30 * time.Second}, // 2^9 = 512, capped at 30
	}

	for _, tt := range tests {
		delay := fetcher.calculateBackoff(tt.attempt)
		if delay != tt.expected {
			t.Errorf("attempt %d: expected delay %v, got %v", tt.attempt, tt.expected, delay)
		}
	}
}

func TestFetcher_InvalidURL(t *testing.T) {
	config := DefaultConfig()
	fetcher := New(config)

	ctx := context.Background()
	_, err := fetcher.Fetch(ctx, "://invalid-url")

	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestFetcher_FollowsRedirects(t *testing.T) {
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final destination"))
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	config := DefaultConfig()
	fetcher := New(config)

	ctx := context.Background()
	resp, err := fetcher.Fetch(ctx, redirectServer.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(resp.Body) != "final destination" {
		t.Errorf("expected 'final destination', got '%s'", resp.Body)
	}
}

func TestFetcher_TooManyRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Infinite redirect loop
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	}))
	defer server.Close()

	config := DefaultConfig()
	fetcher := New(config)

	ctx := context.Background()
	_, err := fetcher.Fetch(ctx, server.URL)

	if err == nil {
		t.Fatal("expected error for too many redirects, got nil")
	}
}
