package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	Config *Config
	Logger *zap.Logger
}

// handleCreate creates a new temporary git repository.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hash, err := GenerateHash()
	if err != nil {
		http.Error(w, "Failed to generate hash", http.StatusInternalServerError)
		s.Logger.Error("Error generating hash", zap.Error(err))
		return
	}

	repoPath := GetRepoPath(hash)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		http.Error(w, "Failed to create repo directory", http.StatusInternalServerError)
		s.Logger.Error("Error creating directory", zap.String("path", repoPath), zap.Error(err))
		return
	}

	// Initialize bare git repository
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, "Failed to initialize git repo", http.StatusInternalServerError)
		s.Logger.Error("Error running git init", zap.String("path", repoPath), zap.Error(err), zap.String("output", string(output)))
		return
	}

	// Enable anonymous push
	cmd = exec.Command("git", "config", "http.receivepack", "true")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, "Failed to configure git repo", http.StatusInternalServerError)
		s.Logger.Error("Error configuring git repo", zap.String("path", repoPath), zap.Error(err), zap.String("output", string(output)))
		return
	}

	url := fmt.Sprintf("https://%s/%s", r.Host, hash)

	s.Logger.Info("Created new repo", zap.String("hash", hash))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", url)
}

// handleGit proxies git requests to git-http-backend.
func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	// Extract the hash from the URL path.
	// Expected format: /<hash>/...
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 1 {
		s.Logger.Debug("Invalid path", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}
	hash := parts[0]

	repoPath := GetRepoPath(hash)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		s.Logger.Debug("Repo not found", zap.String("path", repoPath), zap.String("hash", hash))
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

	s.Logger.Debug("Proxying git request", zap.String("hash", hash), zap.String("path", r.URL.Path))
	handler.ServeHTTP(w, r)
}

func (s *Server) handleTTL(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", s.Config.TTL)
}

func (s *Server) handleRepoTTL(w http.ResponseWriter, r *http.Request) {
	// Expected format: /ttl/<hash>
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/ttl/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Invalid repo hash", http.StatusBadRequest)
		return
	}
	hash := parts[0]

	repoPath := GetRepoPath(hash)
	info, err := os.Stat(repoPath)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		s.Logger.Error("Error stating repo", zap.String("path", repoPath), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(info.ModTime())
	remaining := s.Config.TTL - elapsed
	if remaining < 0 {
		remaining = 0
	}

	fmt.Fprintf(w, "%s\n", remaining)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Expected format: /delete/<hash>
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/delete/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Invalid repo hash", http.StatusBadRequest)
		return
	}
	hash := parts[0]

	repoPath := GetRepoPath(hash)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	if err := os.RemoveAll(repoPath); err != nil {
		s.Logger.Error("Error deleting repo", zap.String("path", repoPath), zap.Error(err))
		http.Error(w, "Failed to delete repo", http.StatusInternalServerError)
		return
	}

	s.Logger.Info("Deleted repo", zap.String("hash", hash))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Repository deleted")
}
