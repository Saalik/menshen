package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

const (
	RepoStore = "repos"
)

func GenerateHash() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetRepoPath(hash string) string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, RepoStore, hash)
}

func EnsureRepoStore() error {
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, RepoStore)
	return os.MkdirAll(path, 0755)
}
