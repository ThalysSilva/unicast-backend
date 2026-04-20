package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count     int
	resetAt   time.Time
	updatedAt time.Time
}

type rateLimiter struct {
	limit   int
	window  time.Duration
	now     func() time.Time
	mu      sync.Mutex
	entries map[string]rateLimitEntry
}

func NewRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &rateLimiter{
		limit:   limit,
		window:  window,
		now:     time.Now,
		entries: make(map[string]rateLimitEntry),
	}

	return limiter.middleware()
}

func (l *rateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l.allow(c) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusTooManyRequests, api.ErrorResponse{
			Error: "Muitas tentativas. Tente novamente em instantes.",
		})
	}
}

func (l *rateLimiter) allow(c *gin.Context) bool {
	now := l.now()
	key := rateLimitKey(c)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanup(now)

	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.resetAt) {
		l.entries[key] = rateLimitEntry{
			count:     1,
			resetAt:   now.Add(l.window),
			updatedAt: now,
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(l.limit-1))
		return true
	}

	if entry.count >= l.limit {
		c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("Retry-After", strconv.Itoa(retryAfterSeconds(now, entry.resetAt)))
		return false
	}

	entry.count++
	entry.updatedAt = now
	l.entries[key] = entry

	c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(l.limit-entry.count))
	return true
}

func (l *rateLimiter) cleanup(now time.Time) {
	for key, entry := range l.entries {
		if now.Sub(entry.updatedAt) > l.window*2 {
			delete(l.entries, key)
		}
	}
}

func rateLimitKey(c *gin.Context) string {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	return c.ClientIP() + ":" + c.Request.Method + ":" + route
}

func retryAfterSeconds(now, resetAt time.Time) int {
	seconds := int(time.Until(resetAt).Seconds())
	if !now.IsZero() {
		seconds = int(resetAt.Sub(now).Seconds())
	}
	if seconds < 1 {
		return 1
	}
	return seconds
}
