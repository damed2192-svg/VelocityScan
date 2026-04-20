package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/yourusername/vps-manager/internal/api"
    "github.com/yourusername/vps-manager/internal/config"
    "github.com/yourusername/vps-manager/internal/docker"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Create Docker client
    dockerClient, err := docker.NewClient(cfg.DockerHost)
    if err != nil {
        log.Fatalf("Failed to create Docker client: %v", err)
    }
    defer dockerClient.Close()
    
    // Check Docker connection
    if err := dockerClient.Ping(); err != nil {
        log.Fatalf("Failed to connect to Docker: %v", err)
    }
    log.Println("Connected to Docker daemon")
    
    // Create data directories
    if err := os.MkdirAll(cfg.DataDir+"/volumes", 0755); err != nil {
        log.Fatalf("Failed to create data directory: %v", err)
    }
    
    // Setup API
    handler := api.NewHandler(dockerClient, cfg)
    router := api.SetupRoutes(handler)
    
    // Start server
    server := &http.Server{
        Addr:    ":" + cfg.APIPort,
        Handler: router,
    }
    
    // Graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        
        log.Println("Shutting down server...")
        if err := server.Close(); err != nil {
            log.Printf("Error closing server: %v", err)
        }
    }()
    
    log.Printf("VPS Manager starting on port %s", cfg.APIPort)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}
