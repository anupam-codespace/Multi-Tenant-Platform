package quickcommerce

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CachedResponse struct {
	StatusCode int
	Body       []byte
	CreatedAt  time.Time
}

// IdempotencyStore stores previous API response bodies keyed by X-Idempotency-Key
type IdempotencyStore struct {
	mu    sync.RWMutex
	cache map[string]CachedResponse
	ttl   time.Duration
}

// NewIdempotencyStore creates a new in-memory response cache for idempotent checkouts
func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	return &IdempotencyStore{
		cache: make(map[string]CachedResponse),
		ttl:   ttl,
	}
}

type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware ensures duplicate requests with identical X-Idempotency-Key return cached response
func (is *IdempotencyStore) IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to state-modifying requests
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
			c.Next()
			return
		}

		key := c.GetHeader("X-Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		is.mu.RLock()
		cached, exists := is.cache[key]
		is.mu.RUnlock()

		if exists && time.Since(cached.CreatedAt) < is.ttl {
			c.Header("X-Cache-Lookup", "HIT-IDEMPOTENT")
			c.Data(cached.StatusCode, "application/json; charset=utf-8", cached.Body)
			c.Abort()
			return
		}

		// Intercept writer to store result
		recorder := &responseRecorder{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = recorder

		c.Next()

		if c.Writer.Status() < 500 {
			is.mu.Lock()
			is.cache[key] = CachedResponse{
				StatusCode: c.Writer.Status(),
				Body:       recorder.body.Bytes(),
				CreatedAt:  time.Now(),
			}
			is.mu.Unlock()
		}
	}
}
