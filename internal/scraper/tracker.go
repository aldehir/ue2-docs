package scraper

import (
	"sync"
	"sync/atomic"
)

// Tracker tracks visited URLs and their HTTP status codes in a thread-safe manner
type Tracker struct {
	visited sync.Map // map[string]int (URL -> status code)
	count   atomic.Int64
}

// NewTracker creates a new URL tracker
func NewTracker() *Tracker {
	return &Tracker{}
}

// MarkVisited marks a URL as visited with the given HTTP status code
func (t *Tracker) MarkVisited(url string, statusCode int) {
	// Check if this is a new URL
	_, existed := t.visited.Swap(url, statusCode)
	if !existed {
		t.count.Add(1)
	}
}

// IsVisited checks if a URL has been visited
func (t *Tracker) IsVisited(url string) bool {
	_, ok := t.visited.Load(url)
	return ok
}

// GetStatus returns the HTTP status code for a visited URL
// Returns (statusCode, true) if the URL has been visited, (0, false) otherwise
func (t *Tracker) GetStatus(url string) (int, bool) {
	val, ok := t.visited.Load(url)
	if !ok {
		return 0, false
	}
	return val.(int), true
}

// VisitedCount returns the total number of unique URLs that have been visited
func (t *Tracker) VisitedCount() int {
	return int(t.count.Load())
}
