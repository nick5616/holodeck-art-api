package adapters

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu             sync.Mutex
	lastSubmission map[string]time.Time
	window         time.Duration
}

func NewRateLimiter(window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		lastSubmission: make(map[string]time.Time),
		window:         window,
	}
	
	// Cleanup old entries every 5 minutes
	go rl.cleanup()
	
	return rl
}

func (r *RateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	lastTime, exists := r.lastSubmission[ip]
	if !exists {
		r.lastSubmission[ip] = time.Now()
		return true
	}
	
	if time.Since(lastTime) < r.window {
		return false
	}
	
	r.lastSubmission[ip] = time.Now()
	return true
}

func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		r.mu.Lock()
		for ip, t := range r.lastSubmission {
			if time.Since(t) > 5*time.Minute {
				delete(r.lastSubmission, ip)
			}
		}
		r.mu.Unlock()
	}
}