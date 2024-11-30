package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var rateLimiters = make(map[string]*RateLimiter)
var rateLimitersMutex sync.Mutex

// GetLimiter retrieves or creates a rate limiter for the given key (e.g., user IP).
func GetLimiter(key string, r rate.Limit, b int) *rate.Limiter {
	rateLimitersMutex.Lock()
	defer rateLimitersMutex.Unlock()

	// Clean up stale rate limiters
	for k, v := range rateLimiters {
		if time.Since(v.lastSeen) > 1*time.Minute {
			delete(rateLimiters, k)
		}
	}

	// Create a new rate limiter if not already present
	if _, exists := rateLimiters[key]; !exists {
		rateLimiters[key] = &RateLimiter{
			limiter:  rate.NewLimiter(r, b), // Rate and burst
			lastSeen: time.Now(),
		}
	}

	// Update the last seen time
	rateLimiters[key].lastSeen = time.Now()
	return rateLimiters[key].limiter
}

// RateLimiterMiddleware creates a middleware function for rate limiting.
func RateLimiterMiddleware(rateLimit rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() // Use IP as the key
		limiter := GetLimiter(key, rateLimit, burst)

		// Check if the request can proceed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
