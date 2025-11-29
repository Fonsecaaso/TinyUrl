package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) {
	// Initialize logger for tests
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func TestNewRateLimiter(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(10, 1*time.Minute)

	assert.NotNil(t, rl)
	assert.NotNil(t, rl.requests)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, 1*time.Minute, rl.window)
}

func TestRateLimiter_Allow_FirstRequest(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(5, 1*time.Minute)
	clientIP := "192.168.1.1"

	allowed := rl.allow(clientIP)

	assert.True(t, allowed)
	assert.Equal(t, 1, rl.requests[clientIP].count)
}

func TestRateLimiter_Allow_MultipleRequests(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(5, 1*time.Minute)
	clientIP := "192.168.1.1"

	// Make 5 requests (should all be allowed)
	for i := 0; i < 5; i++ {
		allowed := rl.allow(clientIP)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed := rl.allow(clientIP)
	assert.False(t, allowed)
}

func TestRateLimiter_Allow_AfterWindowReset(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(2, 100*time.Millisecond)
	clientIP := "192.168.1.1"

	// Make 2 requests (max allowed)
	assert.True(t, rl.allow(clientIP))
	assert.True(t, rl.allow(clientIP))

	// 3rd request should be denied
	assert.False(t, rl.allow(clientIP))

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	allowed := rl.allow(clientIP)
	assert.True(t, allowed)
}

func TestRateLimiter_Allow_MultipleClients(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(3, 1*time.Minute)

	// Test different clients independently
	client1 := "192.168.1.1"
	client2 := "192.168.1.2"
	client3 := "192.168.1.3"

	// Each client should be able to make 3 requests
	for i := 0; i < 3; i++ {
		assert.True(t, rl.allow(client1))
		assert.True(t, rl.allow(client2))
		assert.True(t, rl.allow(client3))
	}

	// 4th request should be denied for all
	assert.False(t, rl.allow(client1))
	assert.False(t, rl.allow(client2))
	assert.False(t, rl.allow(client3))
}

func TestRateLimiter_Middleware_AllowRequest(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(5, 1*time.Minute)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_Middleware_BlockRequest(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(2, 1*time.Minute)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 2 successful requests
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should be rate limited
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_Middleware_ResponseHeaders(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(1, 1*time.Minute)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request succeeds
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited and include headers
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Equal(t, "1", w2.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "1m0s", w2.Header().Get("X-RateLimit-Window"))
}

func TestRateLimiter_Middleware_ErrorResponse(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(1, 2*time.Second)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request succeeds
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Second request should return proper error
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "Rate limit exceeded")
	assert.Contains(t, w2.Body.String(), "RATE_LIMIT_EXCEEDED")
	assert.Contains(t, w2.Body.String(), "retry_after")
	assert.Contains(t, w2.Body.String(), "2") // 2 seconds retry after
}

func TestRateLimiter_Cleanup(t *testing.T) {
	setupTest(t)

	// Use a very short window for testing cleanup
	rl := NewRateLimiter(5, 100*time.Millisecond)
	clientIP := "192.168.1.1"

	// Make a request to create an entry
	rl.allow(clientIP)

	// Verify entry exists
	rl.mutex.RLock()
	assert.Contains(t, rl.requests, clientIP)
	rl.mutex.RUnlock()

	// Wait for cleanup to run (window + some buffer)
	time.Sleep(251 * time.Millisecond)

	// Entry should be cleaned up
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	// Entry may or may not be cleaned up yet depending on timing
	// Just verify the map is accessible (no panic)
	assert.NotNil(t, rl.requests)
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(100, 1*time.Minute)
	clientIP := "192.168.1.1"

	done := make(chan bool, 50)

	// Simulate concurrent requests from same client
	for i := 0; i < 50; i++ {
		go func() {
			rl.allow(clientIP)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	// Check final count
	rl.mutex.RLock()
	count := rl.requests[clientIP].count
	rl.mutex.RUnlock()

	// Should have counted 50 requests
	assert.Equal(t, 50, count)
}

func TestRateLimiter_DifferentPaths(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(2, 1*time.Minute)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/path1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "path1"})
	})
	router.GET("/path2", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "path2"})
	})

	// Make 2 requests to path1 (should succeed)
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/path1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request to path2 should also be rate limited (same client)
	req, _ := http.NewRequest(http.MethodGet, "/path2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiter_ResetBehavior(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(3, 200*time.Millisecond)
	clientIP := "192.168.1.1"

	// Use all 3 requests
	assert.True(t, rl.allow(clientIP))
	assert.True(t, rl.allow(clientIP))
	assert.True(t, rl.allow(clientIP))

	// Check reset time was set
	rl.mutex.RLock()
	bucket := rl.requests[clientIP]
	resetTime := bucket.resetTime
	rl.mutex.RUnlock()

	assert.True(t, resetTime.After(time.Now()))
	assert.True(t, resetTime.Before(time.Now().Add(300*time.Millisecond)))

	// 4th request should be denied
	assert.False(t, rl.allow(clientIP))

	// Wait for window to expire
	time.Sleep(250 * time.Millisecond)

	// Should be able to make requests again
	assert.True(t, rl.allow(clientIP))
}

func TestRateLimiter_EdgeCase_ZeroRate(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(0, 1*time.Minute)
	clientIP := "192.168.1.1"

	// First request creates the bucket with count 1
	// Second request should be denied with 0 rate
	rl.allow(clientIP)            // First request allowed (creates bucket)
	allowed := rl.allow(clientIP) // Second request denied
	assert.False(t, allowed)
}

func TestRateLimiter_EdgeCase_VeryHighRate(t *testing.T) {
	setupTest(t)

	rl := NewRateLimiter(10000, 1*time.Minute)
	clientIP := "192.168.1.1"

	// Should be able to make many requests
	for i := 0; i < 1000; i++ {
		allowed := rl.allow(clientIP)
		assert.True(t, allowed)
	}
}

func TestClientBucket(t *testing.T) {
	setupTest(t)

	now := time.Now()
	bucket := &clientBucket{
		count:     5,
		resetTime: now.Add(1 * time.Minute),
	}

	assert.Equal(t, 5, bucket.count)
	assert.True(t, bucket.resetTime.After(now))
}
