package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	global *rate.Limiter
	repos  map[string]*rate.Limiter
	mu     sync.Mutex
	config RateLimits
	logger *zap.Logger
}

func NewRateLimiter(config RateLimits, logger *zap.Logger) *RateLimiter {
	var global *rate.Limiter
	if config.Global > 0 {
		global = rate.NewLimiter(rate.Limit(config.Global), config.Global)
	}

	return &RateLimiter{
		global: global,
		repos:  make(map[string]*rate.Limiter),
		config: config,
		logger: logger,
	}
}

func (rl *RateLimiter) getRepoLimiter(hash string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.repos[hash]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.config.Repo), rl.config.Repo)
		rl.repos[hash] = limiter
	}

	return limiter
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Global Rate Limit
		if rl.global != nil && !rl.global.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Repo Rate Limit
		// Extract hash from path if applicable
		// Paths: /<hash>/..., /ttl/<hash>, /delete/<hash>
		path := r.URL.Path
		var hash string
		if strings.HasPrefix(path, "/ttl/") {
			parts := strings.Split(strings.TrimPrefix(path, "/ttl/"), "/")
			if len(parts) > 0 {
				hash = parts[0]
			}
		} else if strings.HasPrefix(path, "/delete/") {
			parts := strings.Split(strings.TrimPrefix(path, "/delete/"), "/")
			if len(parts) > 0 {
				hash = parts[0]
			}
		} else if path != "/new" && path != "/metrics" && path != "/ttl" {
			// Assume it's a git request /<hash>/...
			parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
			if len(parts) > 0 {
				hash = parts[0]
			}
		}

		if hash != "" && rl.config.Repo > 0 {
			limiter := rl.getRepoLimiter(hash)
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
		}

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		// Record metrics
		// Use a simplified path for metrics to avoid high cardinality
		metricPath := path
		if hash != "" {
			if strings.HasPrefix(path, "/ttl/") {
				metricPath = "/ttl/:hash"
			} else if strings.HasPrefix(path, "/delete/") {
				metricPath = "/delete/:hash"
			} else {
				metricPath = "/:hash"
			}
		}

		httpRequestsTotal.WithLabelValues(r.Method, metricPath, strconv.Itoa(rw.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, metricPath).Observe(duration)

		rl.logger.Debug("Request processed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
