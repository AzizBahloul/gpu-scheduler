package core

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
)

// Queue manages the scheduling queue with priority
type Queue struct {
	mu       sync.RWMutex
	items    PriorityQueue
	jobMap   map[string]*QueueItem
	maxSize  int
}

// QueueItem represents a job in the queue
type QueueItem struct {
	Job       *models.Job
	Priority  int
	Index     int
	EnqueuedAt time.Time
	AgingBoost int
}

// PriorityQueue implements heap.Interface
type PriorityQueue []*QueueItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority first (max heap)
	totalPriorityI := pq[i].Priority + pq[i].AgingBoost
	totalPriorityJ := pq[j].Priority + pq[j].AgingBoost
	
	if totalPriorityI != totalPriorityJ {
		return totalPriorityI > totalPriorityJ
	}
	
	// If same priority, FIFO
	return pq[i].EnqueuedAt.Before(pq[j].EnqueuedAt)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*QueueItem)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// NewQueue creates a new scheduling queue
func NewQueue(maxSize int) *Queue {
	q := &Queue{
		items:   make(PriorityQueue, 0),
		jobMap:  make(map[string]*QueueItem),
		maxSize: maxSize,
	}
	heap.Init(&q.items)
	return q
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(job *models.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.maxSize {
		return &QueueFullError{MaxSize: q.maxSize}
	}

	if _, exists := q.jobMap[job.ID]; exists {
		return &JobAlreadyInQueueError{JobID: job.ID}
	}

	item := &QueueItem{
		Job:        job,
		Priority:   job.Priority,
		EnqueuedAt: time.Now(),
		AgingBoost: 0,
	}

	heap.Push(&q.items, item)
	q.jobMap[job.ID] = item

	return nil
}

// Dequeue removes and returns the highest priority job
func (q *Queue) Dequeue() *models.Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := heap.Pop(&q.items).(*QueueItem)
	delete(q.jobMap, item.Job.ID)

	return item.Job
}

// Peek returns the highest priority job without removing it
func (q *Queue) Peek() *models.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.items) == 0 {
		return nil
	}

	return q.items[0].Job
}

// Remove removes a specific job from the queue
func (q *Queue) Remove(jobID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.jobMap[jobID]
	if !exists {
		return false
	}

	heap.Remove(&q.items, item.Index)
	delete(q.jobMap, jobID)

	return true
}

// Get returns a job by ID without removing it
func (q *Queue) Get(jobID string) *models.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if item, exists := q.jobMap[jobID]; exists {
		return item.Job
	}
	return nil
}

// Size returns the current queue size
func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// IsEmpty checks if queue is empty
func (q *Queue) IsEmpty() bool {
	return q.Size() == 0
}

// List returns all jobs in the queue
func (q *Queue) List() []*models.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]*models.Job, len(q.items))
	for i, item := range q.items {
		jobs[i] = item.Job
	}
	return jobs
}

// ApplyAging increases priority of waiting jobs to prevent starvation
func (q *Queue) ApplyAging(agingFactor int, ageThreshold time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	for _, item := range q.items {
		waitTime := now.Sub(item.EnqueuedAt)
		if waitTime > ageThreshold {
			// Increase aging boost based on wait time
			item.AgingBoost += agingFactor
		}
	}

	// Re-heapify after priority changes
	heap.Init(&q.items)
}

// GetPosition returns the queue position of a job (1-indexed)
func (q *Queue) GetPosition(jobID string) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, exists := q.jobMap[jobID]
	if !exists {
		return -1
	}

	// Find position in sorted order
	position := 1
	for _, qItem := range q.items {
		if qItem.Job.ID == jobID {
			return position
		}
		totalPriority := qItem.Priority + qItem.AgingBoost
		itemTotalPriority := item.Priority + item.AgingBoost
		
		if totalPriority > itemTotalPriority ||
			(totalPriority == itemTotalPriority && qItem.EnqueuedAt.Before(item.EnqueuedAt)) {
			position++
		}
	}

	return position
}

// Clear removes all jobs from the queue
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = make(PriorityQueue, 0)
	q.jobMap = make(map[string]*QueueItem)
	heap.Init(&q.items)
}

// QueueFullError when queue is at capacity
type QueueFullError struct {
	MaxSize int
}

func (e *QueueFullError) Error() string {
	return fmt.Sprintf("queue is full (max size: %d)", e.MaxSize)
}

// JobAlreadyInQueueError when job is already queued
type JobAlreadyInQueueError struct {
	JobID string
}

func (e *JobAlreadyInQueueError) Error() string {
	return fmt.Sprintf("job %s is already in queue", e.JobID)
}
