package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RateLimiter struct {
	requests map[string]*clientBucket
	mutex    sync.RWMutex
	rate     int
	window   time.Duration
}

type clientBucket struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientBucket),
		rate:     requestsPerWindow,
		window:   window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !rl.allow(clientIP) {
			zap.L().Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
			)
			
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rate))
			c.Header("X-RateLimit-Window", rl.window.String())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	bucket, exists := rl.requests[clientIP]
	
	if !exists || now.After(bucket.resetTime) {
		rl.requests[clientIP] = &clientBucket{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}
	
	if bucket.count >= rl.rate {
		return false
	}
	
	bucket.count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for ip, bucket := range rl.requests {
			if now.After(bucket.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mutex.Unlock()
	}
}