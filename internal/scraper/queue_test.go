package scraper

import (
	"fmt"
	"sync"
	"testing"

	"github.com/aldehir/ue2-docs/internal/urlutil"
)

func TestQueue_AddAndPop(t *testing.T) {
	q := NewQueue()

	// Add items with different weights
	q.Add("https://example.com/page.html", urlutil.ResourceHTML)
	q.Add("https://example.com/image.png", urlutil.ResourceImage)
	q.Add("https://example.com/style.css", urlutil.ResourceCSS)

	// Pop should return highest priority first (HTML > CSS > Image)
	item, ok := q.Pop()
	if !ok {
		t.Fatal("Expected to pop item, but queue was empty")
	}
	if item.URL != "https://example.com/page.html" {
		t.Errorf("Pop() = %v, want HTML resource first", item.URL)
	}

	item, ok = q.Pop()
	if !ok {
		t.Fatal("Expected to pop item, but queue was empty")
	}
	if item.URL != "https://example.com/style.css" {
		t.Errorf("Pop() = %v, want CSS resource second", item.URL)
	}

	item, ok = q.Pop()
	if !ok {
		t.Fatal("Expected to pop item, but queue was empty")
	}
	if item.URL != "https://example.com/image.png" {
		t.Errorf("Pop() = %v, want Image resource third", item.URL)
	}

	// Queue should now be empty
	_, ok = q.Pop()
	if ok {
		t.Error("Expected queue to be empty")
	}
}

func TestQueue_IsEmpty(t *testing.T) {
	q := NewQueue()

	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}

	q.Add("https://example.com/page.html", urlutil.ResourceHTML)

	if q.IsEmpty() {
		t.Error("Queue should not be empty after adding item")
	}

	q.Pop()

	if !q.IsEmpty() {
		t.Error("Queue should be empty after popping all items")
	}
}

func TestQueue_Len(t *testing.T) {
	q := NewQueue()

	if q.Len() != 0 {
		t.Errorf("Len() = %v, want 0", q.Len())
	}

	q.Add("https://example.com/1.html", urlutil.ResourceHTML)
	q.Add("https://example.com/2.html", urlutil.ResourceHTML)
	q.Add("https://example.com/3.html", urlutil.ResourceHTML)

	if q.Len() != 3 {
		t.Errorf("Len() = %v, want 3", q.Len())
	}

	q.Pop()

	if q.Len() != 2 {
		t.Errorf("Len() = %v, want 2 after pop", q.Len())
	}
}

func TestQueue_Deduplication(t *testing.T) {
	q := NewQueue()

	// Add the same URL twice
	added1 := q.Add("https://example.com/page.html", urlutil.ResourceHTML)
	added2 := q.Add("https://example.com/page.html", urlutil.ResourceHTML)

	if !added1 {
		t.Error("First Add() should return true")
	}

	if added2 {
		t.Error("Second Add() for duplicate URL should return false")
	}

	// Queue should only have one item
	if q.Len() != 1 {
		t.Errorf("Len() = %v, want 1 (deduplication should prevent duplicate)", q.Len())
	}
}

func TestQueue_PriorityOrdering(t *testing.T) {
	q := NewQueue()

	// Add items in random order
	q.Add("https://example.com/image.png", urlutil.ResourceImage)    // Weight: 25
	q.Add("https://example.com/page.html", urlutil.ResourceHTML)     // Weight: 100
	q.Add("https://example.com/script.js", urlutil.ResourceJS)       // Weight: 50
	q.Add("https://example.com/style.css", urlutil.ResourceCSS)      // Weight: 75
	q.Add("https://example.com/font.woff", urlutil.ResourceOther)    // Weight: 10

	// Expected resource types in priority order (not checking exact URLs for same-weight items)
	expectedTypes := []urlutil.ResourceType{
		urlutil.ResourceHTML,  // 100
		urlutil.ResourceCSS,   // 75
		urlutil.ResourceJS,    // 50
		urlutil.ResourceImage, // 25
		urlutil.ResourceOther, // 10
	}

	for i, expectedType := range expectedTypes {
		item, ok := q.Pop()
		if !ok {
			t.Fatalf("Pop %d: expected item but queue was empty", i)
		}
		if item.Type != expectedType {
			t.Errorf("Pop %d: got type %v, want %v", i, item.Type, expectedType)
		}
	}
}

func TestQueue_Concurrent(t *testing.T) {
	q := NewQueue()
	numGoroutines := 100
	itemsPerGoroutine := 10

	// Concurrent adds
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				// Use fmt.Sprintf to create unique URLs
				url := fmt.Sprintf("https://example.com/page-%d-%d.html", id, j)
				q.Add(url, urlutil.ResourceHTML)
			}
		}(i)
	}
	wg.Wait()

	expectedLen := numGoroutines * itemsPerGoroutine
	if q.Len() != expectedLen {
		t.Errorf("After concurrent adds, Len() = %v, want %v", q.Len(), expectedLen)
	}

	// Concurrent pops
	poppedCount := 0
	var countMutex sync.Mutex

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				_, ok := q.Pop()
				if ok {
					countMutex.Lock()
					poppedCount++
					countMutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	if poppedCount != expectedLen {
		t.Errorf("Popped %v items, want %v", poppedCount, expectedLen)
	}

	if !q.IsEmpty() {
		t.Error("Queue should be empty after all pops")
	}
}

func TestQueue_PopEmpty(t *testing.T) {
	q := NewQueue()

	item, ok := q.Pop()
	if ok {
		t.Errorf("Pop() on empty queue should return false, got item: %v", item)
	}
}

func TestQueueItem_Weight(t *testing.T) {
	tests := []struct {
		name         string
		resourceType urlutil.ResourceType
		wantWeight   int
	}{
		{"HTML has highest weight", urlutil.ResourceHTML, 100},
		{"CSS has second weight", urlutil.ResourceCSS, 75},
		{"JS has third weight", urlutil.ResourceJS, 50},
		{"Image has fourth weight", urlutil.ResourceImage, 25},
		{"Other has lowest weight", urlutil.ResourceOther, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &QueueItem{
				URL:  "https://example.com/resource",
				Type: tt.resourceType,
			}
			if item.Weight() != tt.wantWeight {
				t.Errorf("Weight() = %v, want %v", item.Weight(), tt.wantWeight)
			}
		})
	}
}
