package scraper

import (
	"sync"
	"testing"
)

func TestTracker_MarkVisited(t *testing.T) {
	tracker := NewTracker()

	url := "https://example.com/page.html"
	statusCode := 200

	// Mark as visited
	tracker.MarkVisited(url, statusCode)

	// Verify it's marked as visited
	if !tracker.IsVisited(url) {
		t.Errorf("Expected URL %q to be marked as visited", url)
	}

	// Verify status code is stored
	code, ok := tracker.GetStatus(url)
	if !ok {
		t.Errorf("Expected to get status code for %q", url)
	}
	if code != statusCode {
		t.Errorf("GetStatus() = %v, want %v", code, statusCode)
	}
}

func TestTracker_IsVisited(t *testing.T) {
	tracker := NewTracker()

	visitedURL := "https://example.com/visited.html"
	unvisitedURL := "https://example.com/unvisited.html"

	tracker.MarkVisited(visitedURL, 200)

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "visited URL returns true",
			url:  visitedURL,
			want: true,
		},
		{
			name: "unvisited URL returns false",
			url:  unvisitedURL,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracker.IsVisited(tt.url)
			if got != tt.want {
				t.Errorf("IsVisited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_GetStatus(t *testing.T) {
	tracker := NewTracker()

	tests := []struct {
		name       string
		url        string
		statusCode int
		setup      func()
		wantCode   int
		wantOk     bool
	}{
		{
			name:       "get status for visited URL",
			url:        "https://example.com/success.html",
			statusCode: 200,
			setup: func() {
				tracker.MarkVisited("https://example.com/success.html", 200)
			},
			wantCode: 200,
			wantOk:   true,
		},
		{
			name:       "get status for 404",
			url:        "https://example.com/notfound.html",
			statusCode: 404,
			setup: func() {
				tracker.MarkVisited("https://example.com/notfound.html", 404)
			},
			wantCode: 404,
			wantOk:   true,
		},
		{
			name:     "get status for unvisited URL",
			url:      "https://example.com/never-visited.html",
			setup:    func() {},
			wantCode: 0,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			code, ok := tracker.GetStatus(tt.url)
			if ok != tt.wantOk {
				t.Errorf("GetStatus() ok = %v, want %v", ok, tt.wantOk)
			}
			if code != tt.wantCode {
				t.Errorf("GetStatus() code = %v, want %v", code, tt.wantCode)
			}
		})
	}
}

func TestTracker_Concurrent(t *testing.T) {
	tracker := NewTracker()

	// Test concurrent writes
	var wg sync.WaitGroup
	numGoroutines := 100
	urlsPerGoroutine := 10

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < urlsPerGoroutine; j++ {
				url := "https://example.com/page-" + string(rune(id)) + "-" + string(rune(j)) + ".html"
				tracker.MarkVisited(url, 200)
			}
		}(i)
	}

	wg.Wait()

	// Verify all URLs were tracked
	count := tracker.VisitedCount()
	expectedCount := numGoroutines * urlsPerGoroutine

	if count != expectedCount {
		t.Errorf("VisitedCount() = %v, want %v", count, expectedCount)
	}
}

func TestTracker_UpdateStatus(t *testing.T) {
	tracker := NewTracker()

	url := "https://example.com/page.html"

	// First visit with 200
	tracker.MarkVisited(url, 200)
	code, _ := tracker.GetStatus(url)
	if code != 200 {
		t.Errorf("First GetStatus() = %v, want 200", code)
	}

	// Update with 404
	tracker.MarkVisited(url, 404)
	code, _ = tracker.GetStatus(url)
	if code != 404 {
		t.Errorf("Second GetStatus() = %v, want 404", code)
	}
}

func TestTracker_VisitedCount(t *testing.T) {
	tracker := NewTracker()

	if tracker.VisitedCount() != 0 {
		t.Errorf("VisitedCount() = %v, want 0", tracker.VisitedCount())
	}

	tracker.MarkVisited("https://example.com/1.html", 200)
	tracker.MarkVisited("https://example.com/2.html", 200)
	tracker.MarkVisited("https://example.com/3.html", 404)

	if tracker.VisitedCount() != 3 {
		t.Errorf("VisitedCount() = %v, want 3", tracker.VisitedCount())
	}

	// Marking the same URL again shouldn't increase count
	tracker.MarkVisited("https://example.com/1.html", 200)

	if tracker.VisitedCount() != 3 {
		t.Errorf("VisitedCount() = %v, want 3 (after duplicate)", tracker.VisitedCount())
	}
}
