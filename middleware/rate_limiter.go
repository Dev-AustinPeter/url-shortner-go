package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Dev-AustinPeter/url-shortner-go/utils"
	"github.com/rs/zerolog"
)

type RateLimiter struct {
	visitors map[string]time.Time
	mutex    sync.Mutex
	limit    time.Duration
	cleanup  time.Duration
	logger   *zerolog.Logger
	stopChan chan struct{}
}

func NewRateLimiter(limit time.Duration, cleanupInterval time.Duration, logger *zerolog.Logger) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]time.Time),
		limit:    limit,
		cleanup:  cleanupInterval,
		logger:   logger,
		stopChan: make(chan struct{}),
	}

	go rl.cleanupExpiredEntries()
	return rl
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		clientIP := rl.getClientIP(r)
		log.Println("clientIP: ", clientIP)

		// Ensures unlocking even if panic occurs
		lastVisit, found := rl.visitors[clientIP]

		if found && time.Since(lastVisit) < rl.limit {
			rl.logger.Warn().Str("ip", clientIP).Msg("Too many requests")
			utils.WriteError(w, http.StatusTooManyRequests, fmt.Errorf("%s", "Too many requests"))
			return
		}

		rl.visitors[clientIP] = time.Now()
		log.Println("clientIP: ", clientIP)
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP, considering proxy headers
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	return r.RemoteAddr
}

// cleanupExpiredEntries periodically removes old entries
func (rl *RateLimiter) cleanupExpiredEntries() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()

			rl.mutex.Lock()
			var expired []string
			for ip, lastVisit := range rl.visitors {
				if now.Sub(lastVisit) > rl.limit {
					expired = append(expired, ip)
				}
			}

			// Remove expired entries
			for _, ip := range expired {
				delete(rl.visitors, ip)
			}
			rl.mutex.Unlock()

		case <-rl.stopChan:
			log.Println("[INFO] Stopping rate limiter cleanup...")
			return
		}
	}
}

// StopCleanup gracefully stops the cleanup goroutine
func (rl *RateLimiter) StopCleanup() {
	close(rl.stopChan)
}
