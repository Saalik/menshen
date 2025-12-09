package main

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	CleanupInterval = 1 * time.Hour
)

// startCleanupTask starts the background cleanup routine.
func startCleanupTask(ttl time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(CleanupInterval)
	go func() {
		for range ticker.C {
			cleanupRepos(ttl, logger)
		}
	}()
}

func cleanupRepos(ttl time.Duration, logger *zap.Logger) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Error getting cwd for cleanup", zap.Error(err))
		return
	}
	repoStorePath := filepath.Join(cwd, RepoStore)

	entries, err := os.ReadDir(repoStorePath)
	if err != nil {
		// If the dir doesn't exist yet, that's fine.
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(repoStorePath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			logger.Error("Error getting info for repo", zap.String("path", repoPath), zap.Error(err))
			continue
		}

		if now.Sub(info.ModTime()) > ttl {
			logger.Info("Deleting expired repo", zap.String("name", entry.Name()))
			if err := os.RemoveAll(repoPath); err != nil {
				logger.Error("Error deleting repo", zap.String("path", repoPath), zap.Error(err))
			}
		}
	}
}
