package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	if err := EnsureRepoStore(); err != nil {
		log.Fatalf("Failed to ensure repo store: %v", err)
	}

	startCleanupTask()

	http.HandleFunc("/new", handleCreate)
	http.HandleFunc("/", handleGit)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Menshen listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
