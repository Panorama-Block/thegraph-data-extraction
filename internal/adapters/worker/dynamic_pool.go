package worker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// DynamicPool implements a worker pool that can adapt its size based on performance metrics
type DynamicPool struct {
	workers       map[int]*worker
	tasks         chan func() error
	wg            sync.WaitGroup
	mu            sync.Mutex
	minWorkers    int
	maxWorkers    int
	currentSize   int32
	idleTimeout   time.Duration
	adjustPeriod  time.Duration
	closed        int32
	taskLatencies []time.Duration
	latencyMu     sync.Mutex
	errorRate     float64
	totalTasks    int64
	successTasks  int64
}

// worker represents a single worker goroutine
type worker struct {
	id         int
	tasks      <-chan func() error
	idle       time.Time
	idleTime   time.Duration
	processing atomic.Bool
	stop       chan struct{}
}

// PoolConfig holds the configuration for the dynamic worker pool
type PoolConfig struct {
	InitialWorkers int
	MinWorkers     int
	MaxWorkers     int
	IdleTimeout    time.Duration
	AdjustPeriod   time.Duration
	QueueSize      int
}

// NewDynamicPool creates a new dynamic worker pool
func NewDynamicPool(config PoolConfig) *DynamicPool {
	// Set defaults for configuration
	if config.InitialWorkers <= 0 {
		config.InitialWorkers = 4
	}
	if config.MinWorkers <= 0 {
		config.MinWorkers = 2
	}
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 20
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 100
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 30 * time.Second
	}
	if config.AdjustPeriod <= 0 {
		config.AdjustPeriod = 5 * time.Second
	}
	
	// Ensure consistent configuration
	if config.MinWorkers > config.InitialWorkers {
		config.InitialWorkers = config.MinWorkers
	}
	if config.MaxWorkers < config.InitialWorkers {
		config.MaxWorkers = config.InitialWorkers
	}
	
	// Create pool
	pool := &DynamicPool{
		workers:      make(map[int]*worker),
		tasks:        make(chan func() error, config.QueueSize),
		minWorkers:   config.MinWorkers,
		maxWorkers:   config.MaxWorkers,
		idleTimeout:  config.IdleTimeout,
		adjustPeriod: config.AdjustPeriod,
		taskLatencies: make([]time.Duration, 0, 100),
	}
	
	// Start initial workers
	for i := 0; i < config.InitialWorkers; i++ {
		pool.startWorker()
	}
	
	// Start the adjustment goroutine
	go pool.adjustWorkers()
	
	return pool
}

// startWorker creates and starts a new worker
func (p *DynamicPool) startWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Find an unused worker ID
	id := 0
	for ; p.workers[id] != nil; id++ {
	}
	
	// Create and start the worker
	w := &worker{
		id:       id,
		tasks:    p.tasks,
		idle:     time.Now(),
		idleTime: p.idleTimeout,
		stop:     make(chan struct{}),
	}
	
	p.workers[id] = w
	atomic.AddInt32(&p.currentSize, 1)
	
	go p.runWorker(w)
	
	log.Debug().
		Int("workerId", id).
		Int("totalWorkers", int(atomic.LoadInt32(&p.currentSize))).
		Msg("Started new worker")
}

// stopWorker stops a worker
func (p *DynamicPool) stopWorker(id int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if w, ok := p.workers[id]; ok {
		// Signal the worker to stop
		close(w.stop)
		delete(p.workers, id)
		atomic.AddInt32(&p.currentSize, -1)
		
		log.Debug().
			Int("workerId", id).
			Int("totalWorkers", int(atomic.LoadInt32(&p.currentSize))).
			Msg("Stopped worker")
	}
}

// runWorker runs a worker until it's stopped
func (p *DynamicPool) runWorker(w *worker) {
	for {
		select {
		case <-w.stop:
			return
		case task, ok := <-w.tasks:
			if !ok {
				return
			}
			
			w.processing.Store(true)
			startTime := time.Now()
			
			// Execute the task
			err := task()
			
			// Record metrics
			latency := time.Since(startTime)
			p.recordTaskCompletion(latency, err == nil)
			
			w.processing.Store(false)
			w.idle = time.Now()
		}
	}
}

// recordTaskCompletion records metrics for a completed task
func (p *DynamicPool) recordTaskCompletion(latency time.Duration, success bool) {
	// Update latency metrics
	p.latencyMu.Lock()
	p.taskLatencies = append(p.taskLatencies, latency)
	
	// Keep only the last 100 latencies
	if len(p.taskLatencies) > 100 {
		p.taskLatencies = p.taskLatencies[len(p.taskLatencies)-100:]
	}
	p.latencyMu.Unlock()
	
	// Update error rate metrics
	atomic.AddInt64(&p.totalTasks, 1)
	if success {
		atomic.AddInt64(&p.successTasks, 1)
	}
	
	// Recalculate error rate
	total := atomic.LoadInt64(&p.totalTasks)
	successful := atomic.LoadInt64(&p.successTasks)
	if total > 0 {
		p.errorRate = 1.0 - float64(successful)/float64(total)
	}
}

// adjustWorkers periodically adjusts the worker pool size based on metrics
func (p *DynamicPool) adjustWorkers() {
	ticker := time.NewTicker(p.adjustPeriod)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&p.closed) != 0 {
				return
			}
			
			p.mu.Lock()
			
			// Check for idle workers to remove
			now := time.Now()
			currentSize := int(atomic.LoadInt32(&p.currentSize))
			
			// Calculate average latency
			avgLatency := p.getAverageLatency()
			errorRate := p.errorRate
			queueSize := len(p.tasks)
			
			// Adaptive scaling logic
			switch {
			case errorRate > 0.25 && currentSize > p.minWorkers:
				// If error rate is high, reduce workers (possible API throttling)
				p.scaleDown(1)
				
			case queueSize > currentSize*2 && currentSize < p.maxWorkers:
				// If queue is filling up, add workers
				p.scaleUp(min(p.maxWorkers-currentSize, 2))
				
			case avgLatency > 2*time.Second && currentSize < p.maxWorkers:
				// If latency is high but error rate is acceptable, add workers
				p.scaleUp(1)
				
			case queueSize == 0 && currentSize > p.minWorkers:
				// If queue is empty, check for idle workers
				for id, w := range p.workers {
					if !w.processing.Load() && now.Sub(w.idle) > w.idleTime && currentSize > p.minWorkers {
						// Stop idle worker
						p.stopWorker(id)
						currentSize--
						if currentSize <= p.minWorkers {
							break
						}
					}
				}
			}
			
			p.mu.Unlock()
		}
	}
}

// scaleUp increases the number of workers
func (p *DynamicPool) scaleUp(count int) {
	currentSize := int(atomic.LoadInt32(&p.currentSize))
	for i := 0; i < count && currentSize+i < p.maxWorkers; i++ {
		p.startWorker()
	}
	
	log.Info().
		Int("newSize", int(atomic.LoadInt32(&p.currentSize))).
		Float64("errorRate", p.errorRate).
		Dur("avgLatency", p.getAverageLatency()).
		Int("queueSize", len(p.tasks)).
		Msg("Scaled up worker pool")
}

// scaleDown decreases the number of workers
func (p *DynamicPool) scaleDown(count int) {
	// Find idle workers to stop
	var workersToStop []int
	
	for id, w := range p.workers {
		if !w.processing.Load() && len(workersToStop) < count {
			workersToStop = append(workersToStop, id)
		}
	}
	
	// Stop the selected workers
	for _, id := range workersToStop {
		p.stopWorker(id)
	}
	
	if len(workersToStop) > 0 {
		log.Info().
			Int("newSize", int(atomic.LoadInt32(&p.currentSize))).
			Float64("errorRate", p.errorRate).
			Dur("avgLatency", p.getAverageLatency()).
			Int("queueSize", len(p.tasks)).
			Msg("Scaled down worker pool")
	}
}

// getAverageLatency calculates the average task latency
func (p *DynamicPool) getAverageLatency() time.Duration {
	p.latencyMu.Lock()
	defer p.latencyMu.Unlock()
	
	if len(p.taskLatencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range p.taskLatencies {
		total += latency
	}
	
	return total / time.Duration(len(p.taskLatencies))
}

// Submit submits a task to the worker pool
func (p *DynamicPool) Submit(task func() error) error {
	if atomic.LoadInt32(&p.closed) != 0 {
		return errors.New("worker pool is closed")
	}
	
	// Submit the task to the queue
	select {
	case p.tasks <- task:
		return nil
	default:
		// If the queue is full, try to add more workers
		currentSize := int(atomic.LoadInt32(&p.currentSize))
		if currentSize < p.maxWorkers {
			p.mu.Lock()
			p.scaleUp(1)
			p.mu.Unlock()
			
			// Try again now that we've added a worker
			select {
			case p.tasks <- task:
				return nil
			default:
				return errors.New("task queue is full")
			}
		}
		return errors.New("task queue is full")
	}
}

// Wait waits for all tasks to complete
func (p *DynamicPool) Wait() error {
	p.wg.Wait()
	return nil
}

// SetPoolSize dynamically adjusts the worker pool size
func (p *DynamicPool) SetPoolSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Bound the size between min and max
	if size < p.minWorkers {
		size = p.minWorkers
	}
	if size > p.maxWorkers {
		size = p.maxWorkers
	}
	
	currentSize := int(atomic.LoadInt32(&p.currentSize))
	
	if size > currentSize {
		// Scale up
		for i := 0; i < size-currentSize; i++ {
			p.startWorker()
		}
	} else if size < currentSize {
		// Scale down
		// Find idle workers to stop
		toStop := currentSize - size
		idleWorkers := make([]int, 0, toStop)
		
		for id, w := range p.workers {
			if !w.processing.Load() && len(idleWorkers) < toStop {
				idleWorkers = append(idleWorkers, id)
			}
		}
		
		// Stop the selected workers
		for _, id := range idleWorkers {
			p.stopWorker(id)
		}
		
		// If we couldn't find enough idle workers, just log it
		if len(idleWorkers) < toStop {
			log.Info().
				Int("requested", size).
				Int("current", currentSize).
				Int("stopped", len(idleWorkers)).
				Msg("Could not stop all requested workers, waiting for tasks to complete")
		}
	}
	
	log.Info().
		Int("oldSize", currentSize).
		Int("newSize", int(atomic.LoadInt32(&p.currentSize))).
		Msg("Pool size adjusted")
}

// Close shuts down the worker pool
func (p *DynamicPool) Close() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return nil // Already closed
	}
	
	// Close the tasks channel to signal workers to exit
	close(p.tasks)
	
	// Stop all workers
	p.mu.Lock()
	for id := range p.workers {
		p.stopWorker(id)
	}
	p.mu.Unlock()
	
	return nil
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 