package ratelimiter

import (
	"context"
	"load-balancer/internal/logger"
	"net"
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	buckets     map[string]*Bucket
	mu          sync.RWMutex
	logger      logger.Logger
	defaultCap  uint32
	defaultRate uint32
}

func New(logger logger.Logger, defaultCap, defaultRate uint32) *Limiter {
	return &Limiter{
		buckets:     map[string]*Bucket{},
		logger:      logger,
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
	}
}

func (r *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		clientID := req.Header.Get("X-Client-ID")
		if clientID == "" {
			clientID = extractIP(req)
		}
		// r.logger.Info("request", "client", clientID)
		if !r.Allow(clientID) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (l *Limiter) GetBucket(clientID string) *Bucket {
	return nil
}

func (l *Limiter) Allow(clientID string) bool {
	l.mu.RLock()
	bucket, ok := l.buckets[clientID]
	l.mu.RUnlock()

	if !ok {
		l.mu.Lock()
		bucket = NewBucket(50, l.defaultCap, l.defaultRate)
		l.buckets[clientID] = bucket
		l.mu.Unlock()
	}

	bucket.RefillIfNeeded(time.Now())

	return bucket.TryTake()
}

func (l *Limiter) RefillAll(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()

			l.mu.RLock()

			for _, bucket := range l.buckets {
				bucket.RefillIfNeeded(now)
			}
			l.mu.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}

func extractIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
