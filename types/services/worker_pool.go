package services

import (
	"container/heap"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/utils"
)

// Service priority levels
const (
	PriorityCritical = 1
	PriorityHigh     = 2
	PriorityNormal   = 3
	PriorityLow      = 4
)

// Default configuration values
const (
	DefaultWorkers            = 50
	DefaultQueueSize          = 1000
	DefaultRateLimitPerDomain = 60
	MinWorkers                = 5
	MaxWorkers                = 500
	MinQueueSize              = 100
	MaxQueueSize              = 10000
)

// checkJob represents a service check to be executed
type checkJob struct {
	service  *Service
	dueTime  time.Time
	priority int
	index    int // heap index
}

// checkJobHeap implements a priority queue for check jobs
// Orders by: priority (ascending), then due time (ascending)
type checkJobHeap []*checkJob

func (h checkJobHeap) Len() int { return len(h) }

func (h checkJobHeap) Less(i, j int) bool {
	// First compare by priority (lower = higher priority)
	if h[i].priority != h[j].priority {
		return h[i].priority < h[j].priority
	}
	// Then by due time
	return h[i].dueTime.Before(h[j].dueTime)
}

func (h checkJobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *checkJobHeap) Push(x interface{}) {
	n := len(*h)
	job := x.(*checkJob)
	job.index = n
	*h = append(*h, job)
}

func (h *checkJobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	job := old[n-1]
	old[n-1] = nil
	job.index = -1
	*h = old[0 : n-1]
	return job
}

// WorkerPool manages concurrent service checks
type WorkerPool struct {
	workers   int
	queueSize int
	rateLimit int

	jobQueue chan *checkJob
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// Scheduler state
	schedulerHeap checkJobHeap
	schedulerMu   sync.Mutex
	schedulerCond *sync.Cond

	// Per-domain rate limiting
	domainLimiter   map[string]*domainBucket
	domainLimiterMu sync.Mutex

	// Metrics
	activeWorkers  int64
	pendingJobs    int64
	completedJobs  int64
	rateLimitedJobs int64

	running atomic.Bool
}

// domainBucket tracks rate limiting per domain
type domainBucket struct {
	tokens    int
	lastReset time.Time
}

var (
	pool   *WorkerPool
	poolMu sync.Mutex
)

// GetWorkerPool returns the singleton worker pool instance
func GetWorkerPool() *WorkerPool {
	poolMu.Lock()
	defer poolMu.Unlock()

	if pool == nil {
		pool = newWorkerPool()
	}
	return pool
}

// ResetWorkerPool resets the singleton for testing purposes
// Should only be called when the pool is stopped
func ResetWorkerPool() {
	poolMu.Lock()
	defer poolMu.Unlock()
	pool = nil
}

// newWorkerPool creates a new worker pool with configuration from Core or env
func newWorkerPool() *WorkerPool {
	workers := DefaultWorkers
	queueSize := DefaultQueueSize
	rateLimit := DefaultRateLimitPerDomain

	// Try to get config from Core
	if core.App != nil {
		if core.App.PollingWorkers > 0 {
			workers = core.App.PollingWorkers
		}
		if core.App.PollingQueueSize > 0 {
			queueSize = core.App.PollingQueueSize
		}
		if core.App.PollingRateLimitPerDomain > 0 {
			rateLimit = core.App.PollingRateLimitPerDomain
		}
	}

	// Override with env vars if set
	if envWorkers := utils.Params.GetInt("POLLING_WORKERS"); envWorkers > 0 {
		workers = envWorkers
	}
	if envQueue := utils.Params.GetInt("POLLING_QUEUE_SIZE"); envQueue > 0 {
		queueSize = envQueue
	}
	if envRate := utils.Params.GetInt("POLLING_RATE_LIMIT"); envRate > 0 {
		rateLimit = envRate
	}

	// Clamp to valid ranges
	if workers < MinWorkers {
		workers = MinWorkers
	} else if workers > MaxWorkers {
		workers = MaxWorkers
	}
	if queueSize < MinQueueSize {
		queueSize = MinQueueSize
	} else if queueSize > MaxQueueSize {
		queueSize = MaxQueueSize
	}

	wp := &WorkerPool{
		workers:       workers,
		queueSize:     queueSize,
		rateLimit:     rateLimit,
		jobQueue:      make(chan *checkJob, queueSize),
		stopCh:        make(chan struct{}),
		schedulerHeap: make(checkJobHeap, 0),
		domainLimiter: make(map[string]*domainBucket),
	}
	wp.schedulerCond = sync.NewCond(&wp.schedulerMu)
	heap.Init(&wp.schedulerHeap)

	log.Infof("Worker pool configured: workers=%d, queue_size=%d, rate_limit=%d/min/domain",
		workers, queueSize, rateLimit)

	return wp
}

// Start initializes and starts the worker pool
func (wp *WorkerPool) Start() {
	if wp.running.Swap(true) {
		return // already running
	}

	log.Infof("Starting worker pool with %d workers", wp.workers)

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Start scheduler
	wp.wg.Add(1)
	go wp.scheduler()

	// Start rate limit cleaner
	wp.wg.Add(1)
	go wp.rateLimitCleaner()
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	if !wp.running.Swap(false) {
		return // not running
	}

	log.Infoln("Stopping worker pool...")
	close(wp.stopCh)

	// Wake up scheduler
	wp.schedulerCond.Broadcast()

	wp.wg.Wait()
	log.Infoln("Worker pool stopped")
}

// ScheduleService adds a service to the scheduler
func (wp *WorkerPool) ScheduleService(s *Service) {
	if !wp.running.Load() {
		return
	}

	wp.schedulerMu.Lock()
	defer wp.schedulerMu.Unlock()

	// Calculate initial due time with stagger
	stagger := time.Duration(s.Id%100) * 50 * time.Millisecond
	dueTime := utils.Now().Add(stagger)

	job := &checkJob{
		service:  s,
		dueTime:  dueTime,
		priority: s.Priority,
	}

	// Validate priority
	if job.priority < PriorityCritical || job.priority > PriorityLow {
		job.priority = PriorityNormal
	}

	heap.Push(&wp.schedulerHeap, job)
	wp.schedulerCond.Signal()
}

// RemoveService removes a service from scheduling
func (wp *WorkerPool) RemoveService(serviceId int64) {
	wp.schedulerMu.Lock()
	defer wp.schedulerMu.Unlock()

	for i, job := range wp.schedulerHeap {
		if job.service.Id == serviceId {
			heap.Remove(&wp.schedulerHeap, i)
			break
		}
	}
}

// UpdateServicePriority updates the priority of a scheduled service
func (wp *WorkerPool) UpdateServicePriority(serviceId int64, newPriority int) {
	wp.schedulerMu.Lock()
	defer wp.schedulerMu.Unlock()

	// Validate priority
	if newPriority < PriorityCritical || newPriority > PriorityLow {
		newPriority = PriorityNormal
	}

	for i, job := range wp.schedulerHeap {
		if job.service.Id == serviceId {
			job.priority = newPriority
			heap.Fix(&wp.schedulerHeap, i)
			break
		}
	}
}

// scheduler runs as a single goroutine managing the timing heap
func (wp *WorkerPool) scheduler() {
	defer wp.wg.Done()

	for {
		wp.schedulerMu.Lock()

		// Wait for jobs or stop signal
		for wp.schedulerHeap.Len() == 0 && wp.running.Load() {
			wp.schedulerCond.Wait()
		}

		if !wp.running.Load() {
			wp.schedulerMu.Unlock()
			return
		}

		// Peek at the next job
		nextJob := wp.schedulerHeap[0]
		now := utils.Now()

		if nextJob.dueTime.After(now) {
			// Not due yet, wait until due time or new job arrives
			waitDur := nextJob.dueTime.Sub(now)
			wp.schedulerMu.Unlock()

			select {
			case <-wp.stopCh:
				return
			case <-time.After(waitDur):
				// Continue to check again
			}
			continue
		}

		// Job is due, pop it
		heap.Pop(&wp.schedulerHeap)
		wp.schedulerMu.Unlock()

		// Check rate limit before queuing
		if wp.checkRateLimit(nextJob.service) {
			// Try to queue the job
			select {
			case wp.jobQueue <- nextJob:
				atomic.AddInt64(&wp.pendingJobs, 1)
			default:
				// Queue full - backpressure, reschedule sooner
				log.Warnf("Job queue full, rescheduling service %s", nextJob.service.Name)
				wp.rescheduleJob(nextJob, 5*time.Second)
			}
		} else {
			// Rate limited, reschedule
			atomic.AddInt64(&wp.rateLimitedJobs, 1)
			// Avoid division by zero - if rate limit is 0 or 1, wait 1 second
			delay := time.Second
			if wp.rateLimit > 1 {
				delay = time.Minute / time.Duration(wp.rateLimit)
			}
			wp.rescheduleJob(nextJob, delay)
		}
	}
}

// rescheduleJob adds a job back to the heap with a new due time
func (wp *WorkerPool) rescheduleJob(job *checkJob, delay time.Duration) {
	wp.schedulerMu.Lock()
	defer wp.schedulerMu.Unlock()

	job.dueTime = utils.Now().Add(delay)
	heap.Push(&wp.schedulerHeap, job)
	wp.schedulerCond.Signal()
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.stopCh:
			return
		case job := <-wp.jobQueue:
			atomic.AddInt64(&wp.activeWorkers, 1)
			atomic.AddInt64(&wp.pendingJobs, -1)

			// Execute the check with panic recovery
			wp.executeCheckSafe(job)

			atomic.AddInt64(&wp.activeWorkers, -1)
			atomic.AddInt64(&wp.completedJobs, 1)

			// Reschedule for next interval (only if still running)
			if wp.running.Load() {
				wp.rescheduleJob(job, job.service.Duration())
			}
		}
	}
}

// executeCheckSafe wraps executeCheck with panic recovery
func (wp *WorkerPool) executeCheckSafe(job *checkJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic in service check for %s: %v", job.service.Name, r)
		}
	}()
	wp.executeCheck(job)
}

// executeCheck performs the actual service check
func (wp *WorkerPool) executeCheck(job *checkJob) {
	s := job.service

	// Skip if service is not running
	if !s.IsRunning() {
		return
	}

	s.CheckService(true)
}

// checkRateLimit returns true if the domain is within rate limits
func (wp *WorkerPool) checkRateLimit(s *Service) bool {
	if wp.rateLimit <= 0 {
		return true // disabled
	}

	domain := extractDomain(s.Domain)
	if domain == "" {
		return true
	}

	wp.domainLimiterMu.Lock()
	defer wp.domainLimiterMu.Unlock()

	bucket, exists := wp.domainLimiter[domain]
	now := time.Now()

	if !exists {
		wp.domainLimiter[domain] = &domainBucket{
			tokens:    wp.rateLimit - 1,
			lastReset: now,
		}
		return true
	}

	// Reset bucket if a minute has passed
	if now.Sub(bucket.lastReset) >= time.Minute {
		bucket.tokens = wp.rateLimit - 1
		bucket.lastReset = now
		return true
	}

	// Check if tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// extractDomain extracts the domain from a URL or address
func extractDomain(addr string) string {
	if addr == "" {
		return ""
	}

	// Try parsing as URL
	if strings.Contains(addr, "://") {
		u, err := url.Parse(addr)
		if err == nil {
			return strings.ToLower(u.Hostname())
		}
	}

	// Handle IPv6 with brackets [::1]:port or [::1]
	if strings.HasPrefix(addr, "[") {
		if idx := strings.Index(addr, "]:"); idx != -1 {
			// [::1]:port format
			return strings.ToLower(addr[1:idx])
		}
		if strings.HasSuffix(addr, "]") {
			// [::1] format (no port)
			return strings.ToLower(addr[1 : len(addr)-1])
		}
	}

	// Check if it looks like IPv6 without brackets (multiple colons)
	if strings.Count(addr, ":") >= 2 && !strings.Contains(addr, "://") {
		// Bare IPv6 address like ::1 or 2001:db8::1
		return strings.ToLower(addr)
	}

	// For TCP/UDP style "host:port" (single colon)
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host := addr[:idx]
		if host == "" {
			return ""
		}
		return strings.ToLower(host)
	}

	return strings.ToLower(addr)
}

// rateLimitCleaner periodically cleans up old rate limit entries
func (wp *WorkerPool) rateLimitCleaner() {
	defer wp.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-wp.stopCh:
			return
		case <-ticker.C:
			wp.cleanRateLimitEntries()
		}
	}
}

func (wp *WorkerPool) cleanRateLimitEntries() {
	wp.domainLimiterMu.Lock()
	defer wp.domainLimiterMu.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for domain, bucket := range wp.domainLimiter {
		if bucket.lastReset.Before(cutoff) {
			delete(wp.domainLimiter, domain)
		}
	}
}

// Stats returns current worker pool statistics
func (wp *WorkerPool) Stats() map[string]interface{} {
	wp.schedulerMu.Lock()
	scheduledCount := wp.schedulerHeap.Len()
	wp.schedulerMu.Unlock()

	return map[string]interface{}{
		"workers":           wp.workers,
		"queue_size":        wp.queueSize,
		"rate_limit":        wp.rateLimit,
		"active_workers":    atomic.LoadInt64(&wp.activeWorkers),
		"pending_jobs":      atomic.LoadInt64(&wp.pendingJobs),
		"scheduled_jobs":    scheduledCount,
		"completed_jobs":    atomic.LoadInt64(&wp.completedJobs),
		"rate_limited_jobs": atomic.LoadInt64(&wp.rateLimitedJobs),
		"running":           wp.running.Load(),
	}
}

// UpdateConfig updates worker pool configuration at runtime
func (wp *WorkerPool) UpdateConfig(workers, queueSize, rateLimit int) {
	// Clamp values
	if workers < MinWorkers {
		workers = MinWorkers
	} else if workers > MaxWorkers {
		workers = MaxWorkers
	}
	if queueSize < MinQueueSize {
		queueSize = MinQueueSize
	} else if queueSize > MaxQueueSize {
		queueSize = MaxQueueSize
	}

	wp.rateLimit = rateLimit

	// Note: Changing workers/queueSize requires restart
	// This is logged as informational
	if workers != wp.workers || queueSize != wp.queueSize {
		log.Infof("Worker pool config updated: workers=%d (current=%d), queue_size=%d (current=%d), rate_limit=%d",
			workers, wp.workers, queueSize, wp.queueSize, rateLimit)
		log.Infoln("Note: Worker count and queue size changes require application restart to take effect")
	}
}
