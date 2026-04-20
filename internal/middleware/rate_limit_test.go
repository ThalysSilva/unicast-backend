package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", NewRateLimiter(2, time.Minute), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/login", nil))
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, recorder.Code, http.StatusNoContent)
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/login", nil))

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatalf("Retry-After header is empty")
	}
}

func TestRateLimiterUsesRouteInKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewRateLimiter(1, time.Minute)
	router := gin.New()
	router.POST("/login", limiter, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.POST("/refresh", limiter, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/login", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/refresh", nil))
	if second.Code != http.StatusNoContent {
		t.Fatalf("refresh status = %d, want %d", second.Code, http.StatusNoContent)
	}
}
