package scraper

import (
	"container/heap"
	"sync"

	"github.com/aldehir/ue2-docs/internal/urlutil"
)

// QueueItem represents an item in the URL queue
type QueueItem struct {
	URL  string
	Type urlutil.ResourceType
}

// Weight returns the priority weight for this item
func (qi *QueueItem) Weight() int {
	return qi.Type.GetWeight()
}

// priorityQueue implements heap.Interface for QueueItem
type priorityQueue []*QueueItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Higher weight = higher priority (so we want descending order)
	return pq[i].Weight() > pq[j].Weight()
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*QueueItem)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// Queue is a thread-safe priority queue for URLs
type Queue struct {
	pq      priorityQueue
	mu      sync.Mutex
	seen    map[string]bool // Track URLs to prevent duplicates
}

// NewQueue creates a new priority queue
func NewQueue() *Queue {
	q := &Queue{
		pq:   make(priorityQueue, 0),
		seen: make(map[string]bool),
	}
	heap.Init(&q.pq)
	return q
}

// Add adds a URL to the queue with the given resource type
// Returns true if the URL was added, false if it was already in the queue
func (q *Queue) Add(url string, resourceType urlutil.ResourceType) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if we've already seen this URL
	if q.seen[url] {
		return false
	}

	// Mark as seen
	q.seen[url] = true

	// Add to priority queue
	item := &QueueItem{
		URL:  url,
		Type: resourceType,
	}
	heap.Push(&q.pq, item)

	return true
}

// Pop removes and returns the highest priority item from the queue
// Returns (item, true) if an item was available, (nil, false) if queue is empty
func (q *Queue) Pop() (*QueueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.pq.Len() == 0 {
		return nil, false
	}

	item := heap.Pop(&q.pq).(*QueueItem)
	return item, true
}

// IsEmpty returns true if the queue is empty
func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len() == 0
}

// Len returns the number of items in the queue
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len()
}
