package main

import (
	"log"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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
	mux.HandleFunc("/", server.handleGit)
	mux.Handle("/metrics", MetricsHandler())

	handler := rateLimiter.Middleware(mux)

	logger.Info("Menshen listening", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
