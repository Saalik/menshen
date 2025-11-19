package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	CleanupInterval = 1 * time.Hour
	RepoTTL         = 48 * time.Hour
)

// startCleanupTask starts the background cleanup routine.
func startCleanupTask() {
	ticker := time.NewTicker(CleanupInterval)
	go func() {
		for range ticker.C {
			cleanupRepos()
		}
	}()
}

func cleanupRepos() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting cwd for cleanup: %v", err)
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
			log.Printf("Error getting info for %s: %v", repoPath, err)
			continue
		}

		if now.Sub(info.ModTime()) > RepoTTL {
			log.Printf("Deleting expired repo: %s", entry.Name())
			if err := os.RemoveAll(repoPath); err != nil {
				log.Printf("Error deleting repo %s: %v", repoPath, err)
			}
		}
	}
}
