package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// AdaptiveLimiter implements a rate limiter that adjusts based on API responses
type AdaptiveLimiter struct {
	limiter      *rate.Limiter
	mu           sync.Mutex
	minRate      float64
	maxRate      float64
	currentRate  float64
	successRate  float64
	latencies    []time.Duration
	latencyIndex int
	latencySize  int
	latencyMu    sync.Mutex
	resetAt      time.Time
	remaining    int
}

// AdaptiveLimiterConfig holds configuration for the adaptive rate limiter
type AdaptiveLimiterConfig struct {
	InitialRate float64
	MinRate     float64
	MaxRate     float64
	Burst       int
}

// NewAdaptiveLimiter creates a new adaptive rate limiter
func NewAdaptiveLimiter(config AdaptiveLimiterConfig) *AdaptiveLimiter {
	// Set defaults
	if config.InitialRate <= 0 {
		config.InitialRate = 5.0 // 5 requests per second by default
	}
	if config.MinRate <= 0 {
		config.MinRate = 1.0 // 1 request per second minimum
	}
	if config.MaxRate <= 0 {
		config.MaxRate = 50.0 // 50 requests per second maximum
	}
	if config.Burst <= 0 {
		config.Burst = 10 // Allow bursts of up to 10 requests
	}
	
	// Ensure consistent configuration
	if config.MinRate > config.InitialRate {
		config.InitialRate = config.MinRate
	}
	if config.MaxRate < config.InitialRate {
		config.MaxRate = config.InitialRate
	}
	
	limiter := &AdaptiveLimiter{
		limiter:     rate.NewLimiter(rate.Limit(config.InitialRate), config.Burst),
		minRate:     config.MinRate,
		maxRate:     config.MaxRate,
		currentRate: config.InitialRate,
		successRate: 1.0,
		latencies:   make([]time.Duration, 100),
		latencySize: 100,
	}
	
	return limiter
}

// Wait blocks until a request is allowed according to rate limits
func (l *AdaptiveLimiter) Wait(ctx context.Context) error {
	// Check if we need to adjust the rate based on reset time
	l.mu.Lock()
	resetAt := l.resetAt
	l.mu.Unlock()
	
	// If we're approaching the reset time and have few requests left, slow down
	if !resetAt.IsZero() && time.Until(resetAt) < 10*time.Second && l.remaining < 10 {
		log.Warn().
			Int("remaining", l.remaining).
			Time("resetAt", resetAt).
			Msg("Approaching API rate limit, reducing rate")
			
		l.reduceRate(0.5) // Reduce rate by half
	}
	
	// Wait according to the current rate
	return l.limiter.Wait(ctx)
}

// Done signals that a request has completed
func (l *AdaptiveLimiter) Done(success bool, latency time.Duration) {
	// Record latency
	l.recordLatency(latency)
	
	// Update success rate
	l.mu.Lock()
	// Use exponential moving average for success rate
	l.successRate = 0.9*l.successRate + 0.1*boolToFloat(success)
	l.mu.Unlock()
	
	// Adjust rate based on success and latency
	l.adjustRate(success, latency)
}

// recordLatency records the latency of a request
func (l *AdaptiveLimiter) recordLatency(latency time.Duration) {
	l.latencyMu.Lock()
	defer l.latencyMu.Unlock()
	
	l.latencies[l.latencyIndex] = latency
	l.latencyIndex = (l.latencyIndex + 1) % l.latencySize
}

// getAverageLatency calculates the average latency of recent requests
func (l *AdaptiveLimiter) getAverageLatency() time.Duration {
	l.latencyMu.Lock()
	defer l.latencyMu.Unlock()
	
	var total time.Duration
	var count int
	
	for _, lat := range l.latencies {
		if lat > 0 {
			total += lat
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / time.Duration(count)
}

// adjustRate adjusts the rate based on success and latency
func (l *AdaptiveLimiter) adjustRate(success bool, latency time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	avgLatency := l.getAverageLatency()
	
	// If requests are failing, reduce the rate
	if !success {
		// Reduce more aggressively if success rate is low
		if l.successRate < 0.7 {
			l.currentRate *= 0.5
		} else {
			l.currentRate *= 0.8
		}
		
		if l.currentRate < l.minRate {
			l.currentRate = l.minRate
		}
		
		log.Info().
			Float64("newRate", l.currentRate).
			Float64("successRate", l.successRate).
			Dur("avgLatency", avgLatency).
			Msg("Reduced rate limit due to failures")
			
		l.limiter.SetLimit(rate.Limit(l.currentRate))
		return
	}
	
	// If latency is high, reduce the rate slightly
	if avgLatency > 500*time.Millisecond {
		l.currentRate *= 0.95
		
		if l.currentRate < l.minRate {
			l.currentRate = l.minRate
		}
		
		log.Debug().
			Float64("newRate", l.currentRate).
			Dur("avgLatency", avgLatency).
			Msg("Reduced rate limit due to high latency")
			
		l.limiter.SetLimit(rate.Limit(l.currentRate))
		return
	}
	
	// If everything looks good, gradually increase the rate
	if l.successRate > 0.95 && avgLatency < 200*time.Millisecond {
		// Increase rate slowly
		l.currentRate *= 1.05
		
		if l.currentRate > l.maxRate {
			l.currentRate = l.maxRate
		}
		
		log.Debug().
			Float64("newRate", l.currentRate).
			Float64("successRate", l.successRate).
			Dur("avgLatency", avgLatency).
			Msg("Increased rate limit due to good performance")
			
		l.limiter.SetLimit(rate.Limit(l.currentRate))
	}
}

// reduceRate reduces the current rate by a factor
func (l *AdaptiveLimiter) reduceRate(factor float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.currentRate *= factor
	
	if l.currentRate < l.minRate {
		l.currentRate = l.minRate
	}
	
	log.Info().
		Float64("newRate", l.currentRate).
		Msg("Manually reduced rate limit")
		
	l.limiter.SetLimit(rate.Limit(l.currentRate))
}

// UpdateRateLimit updates the rate limit based on API response headers
func (l *AdaptiveLimiter) UpdateRateLimit(rateLimit, remaining int, resetAt time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Store values for later use
	l.remaining = remaining
	l.resetAt = resetAt
	
	// If we have a very small number of requests left, reduce rate dramatically
	if remaining < 5 && !resetAt.IsZero() && time.Until(resetAt) > 5*time.Second {
		l.currentRate = l.minRate
		log.Warn().
			Int("remaining", remaining).
			Time("resetAt", resetAt).
			Float64("newRate", l.currentRate).
			Msg("Almost reached API rate limit, setting minimum rate")
	} else if rateLimit > 0 {
		// Set our maximum rate to 80% of the API's rate limit
		suggestedMax := float64(rateLimit) * 0.8
		
		// Only adjust if it's lower than our current max
		if suggestedMax < l.maxRate {
			l.maxRate = suggestedMax
			
			// If current rate exceeds the new max, adjust it
			if l.currentRate > l.maxRate {
				l.currentRate = l.maxRate
			}
			
			log.Info().
				Int("apiLimit", rateLimit).
				Float64("newMaxRate", l.maxRate).
				Float64("currentRate", l.currentRate).
				Msg("Adjusted maximum rate based on API limit")
		}
	}
	
	// Update the limiter with the current rate
	l.limiter.SetLimit(rate.Limit(l.currentRate))
}

// boolToFloat converts a boolean to a float64 (1.0 for true, 0.0 for false)
func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
} 