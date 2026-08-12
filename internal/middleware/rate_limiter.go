package middleware

import (
	"net"
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
		ip := realIP(c)

		val, _ := ipStore.LoadOrStore(ip, &ipRecord{})
		record := val.(*ipRecord)

		record.mu.Lock()
		defer record.mu.Unlock()

		now := time.Now()
		if !record.lastSeen.IsZero() && now.Sub(record.lastSeen) < donationRateLimitDuration {
			retryAfter := int(donationRateLimitDuration.Seconds() - now.Sub(record.lastSeen).Seconds())
			c.Header("Retry-After", "120")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Terlalu banyak permintaan. Silakan tunggu sebelum donasi lagi.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		record.lastSeen = now
		c.Next()
	}
}

// realIP extracts the real client IP, respecting X-Forwarded-For / X-Real-IP headers.
func realIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first non-empty IP in the list
		for _, candidate := range splitCSV(xff) {
			if ip := net.ParseIP(candidate); ip != nil {
				return ip.String()
			}
		}
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
}

func splitCSV(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := trim(s[start:i])
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	return parts
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
