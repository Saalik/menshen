package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	var cfg *Config
	var err error

	configPaths := []string{"config.yaml", "/etc/menshen/config.yaml"}
	for _, path := range configPaths {
		cfg, err = LoadConfig(path)
		if err == nil {
			log.Printf("Loaded config from %s", path)
			break
		}
	}

	if cfg == nil {
		log.Fatalf("Failed to load config from any of the standard paths: %v", err)
	}

	logger, err := InitLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	if err := EnsureRepoStore(); err != nil {
		logger.Fatal("Failed to ensure repo store", zap.Error(err))
	}

	startCleanupTask(cfg.TTL, logger)

	server := &Server{Config: cfg, Logger: logger}
	rateLimiter := NewRateLimiter(cfg.RateLimits, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/ttl", server.handleTTL)
	mux.HandleFunc("/ttl/", server.handleRepoTTL)
	mux.HandleFunc("/delete/", server.handleDelete)
	mux.HandleFunc("/new", server.handleCreate)
	mux.HandleFunc("/new/", server.handleCreate)
	mux.HandleFunc("/", server.handleGit)
	mux.Handle("/metrics", MetricsHandler())

	handler := rateLimiter.Middleware(mux)
	handlerWithCORS := corsMiddleware(handler)

	logger.Info("Menshen listening", zap.String("port", cfg.Port))
	
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handlerWithCORS,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
