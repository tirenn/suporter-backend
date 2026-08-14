package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const donationRateLimitDuration = 2 * time.Minute

type ipRecord struct {
	lastSeen time.Time
	mu       sync.Mutex
}

var (
	ipStore sync.Map // map[string]*ipRecord
)

// DonationRateLimiter limits public POST /donations to 1 request per 2 minutes per IP (with 3s debounce for test calls).
func DonationRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		isTest := c.GetHeader("X-Is-Test") == "true"
		limitDuration := donationRateLimitDuration
		if isTest {
			limitDuration = 3 * time.Second
		}

		ip := c.ClientIP()

		val, _ := ipStore.LoadOrStore(ip, &ipRecord{})
		record := val.(*ipRecord)

		record.mu.Lock()
		defer record.mu.Unlock()

		now := time.Now()
		if !record.lastSeen.IsZero() && now.Sub(record.lastSeen) < limitDuration {
			retryAfter := int(limitDuration.Seconds() - now.Sub(record.lastSeen).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", time.Duration(retryAfter).String())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Terlalu banyak permintaan. Silakan tunggu sebelum mencoba lagi.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		record.lastSeen = now
		c.Next()
	}
}
