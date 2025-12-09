package fetcher

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/aldehir/ue2-docs/internal/urlutil"
)

// Response represents a fetched resource
type Response struct {
	URL          string
	StatusCode   int
	ContentType  string
	ResourceType urlutil.ResourceType
	Body         []byte
	Headers      http.Header
}

// Config holds fetcher configuration
type Config struct {
	Timeout       time.Duration
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	UserAgent     string
	RateLimiter   RateLimiter
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		UserAgent:    "ue2-docs-scraper/1.0",
		RateLimiter:  nil, // No rate limiting by default
	}
}

// Fetcher handles HTTP requests with retry logic and rate limiting
type Fetcher struct {
	client *http.Client
	config Config
}

// New creates a new Fetcher with the given configuration
func New(config Config) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: config.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Follow up to 10 redirects
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		config: config,
	}
}

// Fetch retrieves a resource from the given URL with retry logic
func (f *Fetcher) Fetch(ctx context.Context, url string) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= f.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := f.calculateBackoff(attempt)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to retry
			}
		}

		// Apply rate limiting if configured
		if f.config.RateLimiter != nil {
			if err := f.config.RateLimiter.Wait(ctx); err != nil {
				return nil, err
			}
		}

		resp, err := f.doFetch(ctx, url)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Don't retry on client errors (4xx), only on server errors (5xx) or network errors
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("client error %d: %w", resp.StatusCode, err)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", f.config.MaxRetries, lastErr)
}

// doFetch performs a single HTTP request
func (f *Fetcher) doFetch(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", f.config.UserAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{
			URL:        url,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
		}, fmt.Errorf("reading response body: %w", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Response{
			URL:        url,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
		}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")

	return &Response{
		URL:          url,
		StatusCode:   resp.StatusCode,
		ContentType:  contentType,
		ResourceType: urlutil.DetectResourceType(url, contentType),
		Body:         body,
		Headers:      resp.Header,
	}, nil
}

// calculateBackoff calculates exponential backoff delay
func (f *Fetcher) calculateBackoff(attempt int) time.Duration {
	delay := float64(f.config.InitialDelay) * math.Pow(2, float64(attempt-1))
	delayDuration := time.Duration(delay)

	if delayDuration > f.config.MaxDelay {
		delayDuration = f.config.MaxDelay
	}

	return delayDuration
}

// RateLimiter is an interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// SimpleRateLimiter implements a simple token bucket rate limiter
type SimpleRateLimiter struct {
	ticker *time.Ticker
	tokens chan struct{}
}

// NewSimpleRateLimiter creates a rate limiter that allows 'requests' per 'duration'
func NewSimpleRateLimiter(requests int, duration time.Duration) *SimpleRateLimiter {
	rl := &SimpleRateLimiter{
		ticker: time.NewTicker(duration / time.Duration(requests)),
		tokens: make(chan struct{}, requests),
	}

	// Fill initial tokens
	for i := 0; i < requests; i++ {
		rl.tokens <- struct{}{}
	}

	// Replenish tokens
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Token bucket is full
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available
func (rl *SimpleRateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

// Stop stops the rate limiter
func (rl *SimpleRateLimiter) Stop() {
	rl.ticker.Stop()
}
