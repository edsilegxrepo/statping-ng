package services

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckJobHeap(t *testing.T) {
	h := &checkJobHeap{}
	heap.Init(h)

	// Add jobs with different priorities
	job1 := &checkJob{priority: PriorityNormal, dueTime: time.Now().Add(1 * time.Second)}
	job2 := &checkJob{priority: PriorityCritical, dueTime: time.Now().Add(2 * time.Second)}
	job3 := &checkJob{priority: PriorityLow, dueTime: time.Now().Add(500 * time.Millisecond)}

	heap.Push(h, job1)
	heap.Push(h, job2)
	heap.Push(h, job3)

	// Critical should come first (lowest priority number)
	popped := heap.Pop(h).(*checkJob)
	assert.Equal(t, PriorityCritical, popped.priority)

	// Normal next
	popped = heap.Pop(h).(*checkJob)
	assert.Equal(t, PriorityNormal, popped.priority)

	// Low last
	popped = heap.Pop(h).(*checkJob)
	assert.Equal(t, PriorityLow, popped.priority)
}

func TestCheckJobHeapSamePriority(t *testing.T) {
	h := &checkJobHeap{}
	heap.Init(h)

	now := time.Now()
	job1 := &checkJob{priority: PriorityNormal, dueTime: now.Add(3 * time.Second)}
	job2 := &checkJob{priority: PriorityNormal, dueTime: now.Add(1 * time.Second)}
	job3 := &checkJob{priority: PriorityNormal, dueTime: now.Add(2 * time.Second)}

	heap.Push(h, job1)
	heap.Push(h, job2)
	heap.Push(h, job3)

	// Should come out in due time order (earliest first)
	popped := heap.Pop(h).(*checkJob)
	assert.True(t, popped.dueTime.Equal(now.Add(1*time.Second)))

	popped = heap.Pop(h).(*checkJob)
	assert.True(t, popped.dueTime.Equal(now.Add(2*time.Second)))

	popped = heap.Pop(h).(*checkJob)
	assert.True(t, popped.dueTime.Equal(now.Add(3*time.Second)))
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path", "example.com"},
		{"http://EXAMPLE.COM:8080/path", "example.com"},
		{"https://sub.example.com", "sub.example.com"},
		{"example.com:443", "example.com"},
		{"192.168.1.1:80", "192.168.1.1"},
		{"[::1]:8080", "::1"},
		{"example.com", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkerPoolStats(t *testing.T) {
	wp := &WorkerPool{
		workers:       50,
		queueSize:     1000,
		rateLimit:     60,
		schedulerHeap: make(checkJobHeap, 0),
		schedulerCond: sync.NewCond(&sync.Mutex{}),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)

	stats := wp.Stats()

	assert.Equal(t, 50, stats["workers"])
	assert.Equal(t, 1000, stats["queue_size"])
	assert.Equal(t, 60, stats["rate_limit"])
	assert.Equal(t, int64(0), stats["active_workers"])
	assert.Equal(t, int64(0), stats["pending_jobs"])
	assert.Equal(t, false, stats["running"])
}

func TestPriorityConstants(t *testing.T) {
	// Ensure priority ordering is correct (lower number = higher priority)
	assert.True(t, PriorityCritical < PriorityHigh)
	assert.True(t, PriorityHigh < PriorityNormal)
	assert.True(t, PriorityNormal < PriorityLow)
}

func TestConfigDefaults(t *testing.T) {
	assert.Equal(t, 50, DefaultWorkers)
	assert.Equal(t, 1000, DefaultQueueSize)
	assert.Equal(t, 60, DefaultRateLimitPerDomain)

	// Min/max bounds
	assert.True(t, MinWorkers < DefaultWorkers)
	assert.True(t, DefaultWorkers < MaxWorkers)
	assert.True(t, MinQueueSize < DefaultQueueSize)
	assert.True(t, DefaultQueueSize < MaxQueueSize)
}

func TestRateLimiting(t *testing.T) {
	wp := &WorkerPool{
		rateLimit:     3, // 3 requests per minute
		domainLimiter: make(map[string]*domainBucket),
	}

	service := &Service{Domain: "https://example.com/test"}

	// First 3 should pass
	assert.True(t, wp.checkRateLimit(service), "First request should pass")
	assert.True(t, wp.checkRateLimit(service), "Second request should pass")
	assert.True(t, wp.checkRateLimit(service), "Third request should pass")

	// Fourth should be rate limited
	assert.False(t, wp.checkRateLimit(service), "Fourth request should be rate limited")
}

func TestRateLimitingDisabled(t *testing.T) {
	wp := &WorkerPool{
		rateLimit:     0, // Disabled
		domainLimiter: make(map[string]*domainBucket),
	}

	service := &Service{Domain: "https://example.com/test"}

	// All should pass when rate limiting is disabled
	for i := 0; i < 100; i++ {
		assert.True(t, wp.checkRateLimit(service), "Request %d should pass when rate limiting disabled", i)
	}
}

func TestRateLimitingDifferentDomains(t *testing.T) {
	wp := &WorkerPool{
		rateLimit:     2,
		domainLimiter: make(map[string]*domainBucket),
	}

	service1 := &Service{Domain: "https://example.com/test"}
	service2 := &Service{Domain: "https://other.com/test"}

	// Each domain has its own bucket
	assert.True(t, wp.checkRateLimit(service1))
	assert.True(t, wp.checkRateLimit(service1))
	assert.False(t, wp.checkRateLimit(service1)) // Rate limited

	// Different domain should still work
	assert.True(t, wp.checkRateLimit(service2))
	assert.True(t, wp.checkRateLimit(service2))
	assert.False(t, wp.checkRateLimit(service2)) // Rate limited
}

func TestRateLimitReset(t *testing.T) {
	wp := &WorkerPool{
		rateLimit:     1,
		domainLimiter: make(map[string]*domainBucket),
	}

	service := &Service{Domain: "https://example.com/test"}

	// First request passes
	assert.True(t, wp.checkRateLimit(service))
	// Second is rate limited
	assert.False(t, wp.checkRateLimit(service))

	// Manually reset the bucket's lastReset to simulate time passing
	wp.domainLimiterMu.Lock()
	wp.domainLimiter["example.com"].lastReset = time.Now().Add(-2 * time.Minute)
	wp.domainLimiterMu.Unlock()

	// Should pass now
	assert.True(t, wp.checkRateLimit(service))
}

func TestWorkerPoolStartStop(t *testing.T) {
	ResetWorkerPool()

	wp := &WorkerPool{
		workers:       2,
		queueSize:     10,
		rateLimit:     60,
		jobQueue:      make(chan *checkJob, 10),
		stopCh:        make(chan struct{}),
		schedulerHeap: make(checkJobHeap, 0),
		domainLimiter: make(map[string]*domainBucket),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)

	assert.False(t, wp.running.Load())

	wp.Start()
	assert.True(t, wp.running.Load())

	// Starting again should be a no-op
	wp.Start()
	assert.True(t, wp.running.Load())

	wp.Stop()
	assert.False(t, wp.running.Load())

	// Stopping again should be a no-op
	wp.Stop()
	assert.False(t, wp.running.Load())
}

func TestScheduleService(t *testing.T) {
	wp := &WorkerPool{
		workers:       2,
		queueSize:     10,
		rateLimit:     60,
		jobQueue:      make(chan *checkJob, 10),
		stopCh:        make(chan struct{}),
		schedulerHeap: make(checkJobHeap, 0),
		domainLimiter: make(map[string]*domainBucket),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)
	wp.running.Store(true)

	service := &Service{Id: 1, Name: "Test", Priority: PriorityHigh}
	wp.ScheduleService(service)

	wp.schedulerMu.Lock()
	assert.Equal(t, 1, wp.schedulerHeap.Len())
	job := wp.schedulerHeap[0]
	assert.Equal(t, PriorityHigh, job.priority)
	assert.Equal(t, int64(1), job.service.Id)
	wp.schedulerMu.Unlock()
}

func TestScheduleServiceInvalidPriority(t *testing.T) {
	wp := &WorkerPool{
		workers:       2,
		queueSize:     10,
		rateLimit:     60,
		jobQueue:      make(chan *checkJob, 10),
		stopCh:        make(chan struct{}),
		schedulerHeap: make(checkJobHeap, 0),
		domainLimiter: make(map[string]*domainBucket),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)
	wp.running.Store(true)

	// Invalid priority (too low)
	service := &Service{Id: 1, Name: "Test", Priority: 0}
	wp.ScheduleService(service)

	wp.schedulerMu.Lock()
	job := wp.schedulerHeap[0]
	assert.Equal(t, PriorityNormal, job.priority, "Invalid priority should default to Normal")
	wp.schedulerMu.Unlock()
}

func TestScheduleServiceWhenNotRunning(t *testing.T) {
	wp := &WorkerPool{
		schedulerHeap: make(checkJobHeap, 0),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)
	wp.running.Store(false)

	service := &Service{Id: 1, Name: "Test", Priority: PriorityNormal}
	wp.ScheduleService(service)

	wp.schedulerMu.Lock()
	assert.Equal(t, 0, wp.schedulerHeap.Len(), "Should not schedule when not running")
	wp.schedulerMu.Unlock()
}

func TestRemoveService(t *testing.T) {
	wp := &WorkerPool{
		schedulerHeap: make(checkJobHeap, 0),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)
	wp.running.Store(true)

	// Add some services
	for i := int64(1); i <= 3; i++ {
		job := &checkJob{
			service:  &Service{Id: i, Name: "Test"},
			priority: PriorityNormal,
			dueTime:  time.Now(),
		}
		heap.Push(&wp.schedulerHeap, job)
	}

	assert.Equal(t, 3, wp.schedulerHeap.Len())

	// Remove service 2
	wp.RemoveService(2)

	assert.Equal(t, 2, wp.schedulerHeap.Len())

	// Verify service 2 is gone
	wp.schedulerMu.Lock()
	for _, job := range wp.schedulerHeap {
		assert.NotEqual(t, int64(2), job.service.Id)
	}
	wp.schedulerMu.Unlock()
}

func TestUpdateServicePriority(t *testing.T) {
	wp := &WorkerPool{
		schedulerHeap: make(checkJobHeap, 0),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)

	// Add a service with Normal priority
	job := &checkJob{
		service:  &Service{Id: 1, Name: "Test"},
		priority: PriorityNormal,
		dueTime:  time.Now(),
	}
	heap.Push(&wp.schedulerHeap, job)

	// Update to Critical
	wp.UpdateServicePriority(1, PriorityCritical)

	wp.schedulerMu.Lock()
	assert.Equal(t, PriorityCritical, wp.schedulerHeap[0].priority)
	wp.schedulerMu.Unlock()
}

func TestUpdateServicePriorityInvalidValue(t *testing.T) {
	wp := &WorkerPool{
		schedulerHeap: make(checkJobHeap, 0),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)

	job := &checkJob{
		service:  &Service{Id: 1, Name: "Test"},
		priority: PriorityCritical,
		dueTime:  time.Now(),
	}
	heap.Push(&wp.schedulerHeap, job)

	// Invalid priority should default to Normal
	wp.UpdateServicePriority(1, 99)

	wp.schedulerMu.Lock()
	assert.Equal(t, PriorityNormal, wp.schedulerHeap[0].priority)
	wp.schedulerMu.Unlock()
}

func TestUpdateConfig(t *testing.T) {
	wp := &WorkerPool{
		workers:   50,
		queueSize: 1000,
		rateLimit: 60,
	}

	// Update rate limit (takes effect immediately)
	wp.UpdateConfig(50, 1000, 100)
	assert.Equal(t, 100, wp.rateLimit)

	// Clamping to min values
	wp.UpdateConfig(1, 10, 30)
	assert.Equal(t, 30, wp.rateLimit) // Rate limit updated

	// Clamping to max values
	wp.UpdateConfig(1000, 50000, 30)
	assert.Equal(t, 30, wp.rateLimit)
}

func TestCleanRateLimitEntries(t *testing.T) {
	wp := &WorkerPool{
		domainLimiter: make(map[string]*domainBucket),
	}

	// Add some entries
	now := time.Now()
	wp.domainLimiter["recent.com"] = &domainBucket{tokens: 5, lastReset: now}
	wp.domainLimiter["old.com"] = &domainBucket{tokens: 5, lastReset: now.Add(-10 * time.Minute)}
	wp.domainLimiter["ancient.com"] = &domainBucket{tokens: 5, lastReset: now.Add(-1 * time.Hour)}

	wp.cleanRateLimitEntries()

	// Only recent.com should remain
	assert.Len(t, wp.domainLimiter, 1)
	assert.Contains(t, wp.domainLimiter, "recent.com")
}

func TestHeapFixAfterPriorityUpdate(t *testing.T) {
	h := &checkJobHeap{}
	heap.Init(h)

	now := time.Now()
	job1 := &checkJob{service: &Service{Id: 1}, priority: PriorityLow, dueTime: now}
	job2 := &checkJob{service: &Service{Id: 2}, priority: PriorityNormal, dueTime: now}
	job3 := &checkJob{service: &Service{Id: 3}, priority: PriorityHigh, dueTime: now}

	heap.Push(h, job1)
	heap.Push(h, job2)
	heap.Push(h, job3)

	// Top should be High priority (id=3)
	assert.Equal(t, int64(3), (*h)[0].service.Id)

	// Change job1 (Low) to Critical
	job1.priority = PriorityCritical
	heap.Fix(h, job1.index)

	// Now top should be job1 (Critical)
	assert.Equal(t, int64(1), (*h)[0].service.Id)
}

func TestConcurrentRateLimiting(t *testing.T) {
	wp := &WorkerPool{
		rateLimit:     100,
		domainLimiter: make(map[string]*domainBucket),
	}

	service := &Service{Domain: "https://example.com/test"}

	var wg sync.WaitGroup
	var passed int64

	// 200 concurrent requests
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if wp.checkRateLimit(service) {
				atomic.AddInt64(&passed, 1)
			}
		}()
	}

	wg.Wait()

	// Should be exactly 100 (the rate limit)
	assert.Equal(t, int64(100), passed)
}

func TestExecuteCheckSafePanicRecovery(t *testing.T) {
	wp := &WorkerPool{}

	// Create a job that will cause a panic (nil service methods)
	job := &checkJob{
		service: &Service{Name: "PanicTest"},
	}

	// This should not panic - it should recover
	require.NotPanics(t, func() {
		wp.executeCheckSafe(job)
	})
}

func TestExtractDomainEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"just port", ":8080", ""},
		{"IPv6 no port", "::1", "::1"},
		{"IPv6 with brackets no port", "[::1]", "::1"},
		{"IPv6 with brackets and port", "[::1]:8080", "::1"},
		{"URL with username", "https://user:pass@example.com/path", "example.com"},
		{"URL with fragment", "https://example.com/path#section", "example.com"},
		{"URL with query", "https://example.com/path?query=1", "example.com"},
		{"bare IPv6", "2001:db8::1", "2001:db8::1"},
		{"full IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkerPoolResetSingleton(t *testing.T) {
	ResetWorkerPool()

	wp1 := GetWorkerPool()
	require.NotNil(t, wp1)

	ResetWorkerPool()

	wp2 := GetWorkerPool()
	require.NotNil(t, wp2)

	// They should be different instances
	assert.NotSame(t, wp1, wp2)
}
