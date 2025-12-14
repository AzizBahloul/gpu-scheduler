package core

import (
	"testing"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue(100)
	assert.NotNil(t, q)
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())
}

func TestEnqueueDequeue(t *testing.T) {
	q := NewQueue(10)

	job := &models.Job{
		ID:       "job-1",
		Name:     "test-job",
		Priority: 100,
		GPUCount: 1,
	}

	err := q.Enqueue(job)
	assert.NoError(t, err)
	assert.Equal(t, 1, q.Size())
	assert.False(t, q.IsEmpty())

	dequeued := q.Dequeue()
	assert.Equal(t, job.ID, dequeued.ID)
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())
}

func TestPriorityOrdering(t *testing.T) {
	q := NewQueue(10)

	jobs := []*models.Job{
		{ID: "job-low", Priority: 100, GPUCount: 1},
		{ID: "job-high", Priority: 1000, GPUCount: 1},
		{ID: "job-medium", Priority: 500, GPUCount: 1},
	}

	for _, job := range jobs {
		err := q.Enqueue(job)
		require.NoError(t, err)
	}

	// Should dequeue in priority order: high, medium, low
	assert.Equal(t, "job-high", q.Dequeue().ID)
	assert.Equal(t, "job-medium", q.Dequeue().ID)
	assert.Equal(t, "job-low", q.Dequeue().ID)
}

func TestFIFOWithinSamePriority(t *testing.T) {
	q := NewQueue(10)

	job1 := &models.Job{ID: "job-1", Priority: 100, GPUCount: 1}
	job2 := &models.Job{ID: "job-2", Priority: 100, GPUCount: 1}
	job3 := &models.Job{ID: "job-3", Priority: 100, GPUCount: 1}

	time.Sleep(10 * time.Millisecond)
	q.Enqueue(job1)
	time.Sleep(10 * time.Millisecond)
	q.Enqueue(job2)
	time.Sleep(10 * time.Millisecond)
	q.Enqueue(job3)

	// Should dequeue in FIFO order
	assert.Equal(t, "job-1", q.Dequeue().ID)
	assert.Equal(t, "job-2", q.Dequeue().ID)
	assert.Equal(t, "job-3", q.Dequeue().ID)
}

func TestQueueMaxSize(t *testing.T) {
	q := NewQueue(2)

	job1 := &models.Job{ID: "job-1", Priority: 100, GPUCount: 1}
	job2 := &models.Job{ID: "job-2", Priority: 100, GPUCount: 1}
	job3 := &models.Job{ID: "job-3", Priority: 100, GPUCount: 1}

	assert.NoError(t, q.Enqueue(job1))
	assert.NoError(t, q.Enqueue(job2))

	err := q.Enqueue(job3)
	assert.Error(t, err)
	assert.Equal(t, 2, q.Size())
}

func TestRemoveJob(t *testing.T) {
	q := NewQueue(10)

	jobs := []*models.Job{
		{ID: "job-1", Priority: 100, GPUCount: 1},
		{ID: "job-2", Priority: 200, GPUCount: 1},
		{ID: "job-3", Priority: 300, GPUCount: 1},
	}

	for _, job := range jobs {
		q.Enqueue(job)
	}

	removed := q.Remove("job-2")
	assert.True(t, removed)
	assert.Equal(t, 2, q.Size())

	// Verify job-2 is not in queue
	assert.Nil(t, q.Get("job-2"))
}

func TestPeek(t *testing.T) {
	q := NewQueue(10)

	job := &models.Job{ID: "job-1", Priority: 100, GPUCount: 1}
	q.Enqueue(job)

	peeked := q.Peek()
	assert.Equal(t, job.ID, peeked.ID)
	assert.Equal(t, 1, q.Size()) // Size should not change
}

func TestGetPosition(t *testing.T) {
	q := NewQueue(10)

	jobs := []*models.Job{
		{ID: "job-high", Priority: 1000, GPUCount: 1},
		{ID: "job-medium", Priority: 500, GPUCount: 1},
		{ID: "job-low", Priority: 100, GPUCount: 1},
	}

	for _, job := range jobs {
		q.Enqueue(job)
	}

	assert.Equal(t, 1, q.GetPosition("job-high"))
	assert.Equal(t, 2, q.GetPosition("job-medium"))
	assert.Equal(t, 3, q.GetPosition("job-low"))
	assert.Equal(t, -1, q.GetPosition("nonexistent"))
}

func TestApplyAging(t *testing.T) {
	q := NewQueue(10)

	oldJob := &models.Job{ID: "old-job", Priority: 100, GPUCount: 1}
	q.Enqueue(oldJob)

	// Manually set old enqueue time
	q.mu.Lock()
	item := q.jobMap["old-job"]
	item.EnqueuedAt = time.Now().Add(-10 * time.Minute)
	q.mu.Unlock()

	newJob := &models.Job{ID: "new-job", Priority: 200, GPUCount: 1}
	q.Enqueue(newJob)

	// Apply aging (boost jobs older than 5 minutes by 150 priority)
	q.ApplyAging(150, 5*time.Minute)

	// Old job with boost (100 + 150 = 250) should now be higher priority than new job (200)
	first := q.Dequeue()
	assert.Equal(t, "old-job", first.ID)
}

func TestClear(t *testing.T) {
	q := NewQueue(10)

	for i := 0; i < 5; i++ {
		q.Enqueue(&models.Job{
			ID:       "job-" + string(rune(i)),
			Priority: 100,
			GPUCount: 1,
		})
	}

	assert.Equal(t, 5, q.Size())

	q.Clear()
	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())
}

func TestListJobs(t *testing.T) {
	q := NewQueue(10)

	jobs := []*models.Job{
		{ID: "job-1", Priority: 100, GPUCount: 1},
		{ID: "job-2", Priority: 200, GPUCount: 1},
		{ID: "job-3", Priority: 300, GPUCount: 1},
	}

	for _, job := range jobs {
		q.Enqueue(job)
	}

	listed := q.List()
	assert.Equal(t, 3, len(listed))
}

func TestConcurrentAccess(t *testing.T) {
	q := NewQueue(100)
	done := make(chan bool)

	// Concurrent enqueues
	for i := 0; i < 10; i++ {
		go func(id int) {
			job := &models.Job{
				ID:       "job-" + string(rune(id)),
				Priority: id * 10,
				GPUCount: 1,
			}
			q.Enqueue(job)
			done <- true
		}(i)
	}

	// Wait for all enqueues
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, q.Size())

	// Concurrent dequeues
	for i := 0; i < 5; i++ {
		go func() {
			q.Dequeue()
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	assert.Equal(t, 5, q.Size())
}

func BenchmarkEnqueue(b *testing.B) {
	q := NewQueue(100000)
	job := &models.Job{ID: "bench-job", Priority: 100, GPUCount: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job.ID = "job-" + string(rune(i))
		q.Enqueue(job)
	}
}

func BenchmarkDequeue(b *testing.B) {
	q := NewQueue(100000)

	// Pre-populate queue
	for i := 0; i < b.N; i++ {
		q.Enqueue(&models.Job{
			ID:       "job-" + string(rune(i)),
			Priority: i,
			GPUCount: 1,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Dequeue()
	}
}
