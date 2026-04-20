package config

import (
    "os"
    "path/filepath"
)

type Config struct {
    DockerHost    string
    DataDir       string
    APIPort       string
    DefaultImage  string
}

func Load() *Config {
    return &Config{
        DockerHost:   getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
        DataDir:      getEnv("DATA_DIR", "./data"),
        APIPort:      getEnv("API_PORT", "8080"),
        DefaultImage: getEnv("DEFAULT_IMAGE", "ubuntu:latest"),
    }
}

func (c *Config) GetVolumePath(vpsID string) string {
    return filepath.Join(c.DataDir, "volumes", vpsID)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
