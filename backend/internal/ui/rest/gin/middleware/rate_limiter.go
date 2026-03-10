package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"phone-call-receptionist/backend/pkg/dtos/responses"
)

type rateBucket struct {
	tokens    int
	lastReset time.Time
}

// RateLimiter limits requests per IP using a simple token bucket.
func RateLimiter(requestsPerMinute int) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*rateBucket)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		bucket, ok := buckets[ip]
		if !ok {
			bucket = &rateBucket{tokens: requestsPerMinute, lastReset: time.Now()}
			buckets[ip] = bucket
		}

		if time.Since(bucket.lastReset) > time.Minute {
			bucket.tokens = requestsPerMinute
			bucket.lastReset = time.Now()
		}

		if bucket.tokens <= 0 {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				responses.ErrorResponse{Error: "rate limit exceeded"})
			return
		}

		bucket.tokens--
		mu.Unlock()

		c.Next()
	}
}
