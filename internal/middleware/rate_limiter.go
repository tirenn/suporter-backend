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

// DonationRateLimiter limits POST /donations to 1 request per 2 minutes per IP.
func DonationRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		val, _ := ipStore.LoadOrStore(ip, &ipRecord{})
		record := val.(*ipRecord)

		record.mu.Lock()
		defer record.mu.Unlock()

		now := time.Now()
		if !record.lastSeen.IsZero() && now.Sub(record.lastSeen) < donationRateLimitDuration {
			retryAfter := int(donationRateLimitDuration.Seconds() - now.Sub(record.lastSeen).Seconds())
			c.Header("Retry-After", "120")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak permintaan. Silakan tunggu sebelum donasi lagi.", "retry_after": retryAfter})
			c.Abort()
			return
		}

		record.lastSeen = now
		c.Next()
	}
}
