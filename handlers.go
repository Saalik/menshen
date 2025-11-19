package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// handleCreate creates a new temporary git repository.
func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hash, err := GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusInternalServerError)
		log.Printf("Error generating hash: %v", err)
		return
	}

	repoPath := GetRepoPath(hash)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		http.Error(w, "Failed to create repo directory", http.StatusInternalServerError)
		log.Printf("Error creating directory %s: %v", repoPath, err)
		return
	}

	// Initialize bare git repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, "Failed to initialize git repo", http.StatusInternalServerError)
		log.Printf("Error running git init in %s: %v, output: %s", repoPath, err, string(output))
		return
	}

	// Enable anonymous push
	cmd = exec.Command("git", "config", "http.receivepack", "true")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, "Failed to configure git repo", http.StatusInternalServerError)
		log.Printf("Error configuring git repo in %s: %v, output: %s", repoPath, err, string(output))
		return
	}

	url := fmt.Sprintf("https://%s/%s", r.Host, hash)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", url)
}

// handleGit proxies git requests to git-http-backend.
func handleGit(w http.ResponseWriter, r *http.Request) {
	// Extract the hash from the URL path.
	// Expected format: /<hash>/...
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 1 {
		log.Printf("Invalid path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	hash := parts[0]

	repoPath := GetRepoPath(hash)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Printf("Repo not found: %s (hash: %s)", repoPath, hash)
		http.NotFound(w, r)
		return
	}

	// Update the last modification time of the repo
	now := time.Now()
	os.Chtimes(repoPath, now, now)

	// Set up CGI handler
	handler := &cgi.Handler{
		Path: "/usr/lib/git-core/git-http-backend",
		Env: []string{
			"GIT_PROJECT_ROOT=" + filepath.Dir(repoPath),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}

	// git-http-backend expects PATH_INFO to start with the repo name.
	// Our URL is /<hash>/info/refs...
	// The repo is located at <RepoStore>/<hash>
	// If GIT_PROJECT_ROOT is <RepoStore>, then PATH_INFO should be /<hash>/info/refs...
	// This matches r.URL.Path.

	// However, we need to ensure GIT_PROJECT_ROOT is absolute.
	cwd, _ := os.Getwd()
	absRepoStore := filepath.Join(cwd, RepoStore)
	handler.Env[0] = "GIT_PROJECT_ROOT=" + absRepoStore

	handler.ServeHTTP(w, r)
}
