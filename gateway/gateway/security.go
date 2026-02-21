// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ReplayProtectionStore struct {
	mu         sync.Mutex
	seenNonces map[string]time.Time
	ttl        time.Duration
}

var GlobalReplayStore *ReplayProtectionStore

func InitSecurity() {
	GlobalReplayStore = &ReplayProtectionStore{
		seenNonces: make(map[string]time.Time),
		ttl:        60 * time.Second,
	}
	go GlobalReplayStore.cleanupLoop()
}

func (s *ReplayProtectionStore) IsReplay(nonce string, timestamp int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqTime := time.Unix(timestamp, 0)
	if time.Since(reqTime) > s.ttl {
		log.Printf("   [SECURITY] Timestamp too old: %v", time.Since(reqTime))
		return true
	}
	if time.Until(reqTime) > s.ttl {
		log.Printf("   [SECURITY] Timestamp in future: %v", time.Until(reqTime))
		return true
	}

	if _, exists := s.seenNonces[nonce]; exists {
		log.Printf("   [SECURITY] Duplicate nonce")
		return true
	}

	s.seenNonces[nonce] = time.Now()
	return false
}

func (s *ReplayProtectionStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.cleanup()
	}
}

func (s *ReplayProtectionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for nonce, seenAt := range s.seenNonces {
		if now.Sub(seenAt) > s.ttl {
			delete(s.seenNonces, nonce)
		}
	}
}

// 100 req/min per IP with periodic cleanup
func RateLimitMiddleware() gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string][]time.Time)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, timestamps := range visitors {
				valid := timestamps[:0]
				for _, t := range timestamps {
					if now.Sub(t) < 1*time.Minute {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(visitors, ip)
				} else {
					visitors[ip] = valid
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.ClientIP()
		}

		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		timestamps := visitors[ip]

		validTimestamps := timestamps[:0]
		for _, t := range timestamps {
			if now.Sub(t) < 1*time.Minute {
				validTimestamps = append(validTimestamps, t)
			}
		}

		if len(validTimestamps) >= 100 {
			log.Printf("   [SECURITY] Rate limit exceeded: %s", ip)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		validTimestamps = append(validTimestamps, now)
		visitors[ip] = validTimestamps

		c.Next()
	}
}
