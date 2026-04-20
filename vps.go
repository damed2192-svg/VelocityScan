package models

import "time"

type VPS struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Image       string            `json:"image"`
    Ports       map[string]string `json:"ports"` // host:container
    Env         []string          `json:"env"`
    CPUQuota    int64             `json:"cpu_quota"`    // micro-seconds per 100ms
    MemoryLimit int64             `json:"memory_limit"` // bytes
    Status      string            `json:"status"`
    CreatedAt   time.Time         `json:"created_at"`
    ContainerID string            `json:"container_id,omitempty"`
}

type CreateVPSRequest struct {
    Name        string            `json:"name"`
    Image       string            `json:"image,omitempty"`
    Ports       map[string]string `json:"ports,omitempty"`
    Env         []string          `json:"env,omitempty"`
    CPUQuota    int64             `json:"cpu_quota,omitempty"`
    MemoryLimit int64             `json:"memory_limit,omitempty"`
}

type ExecCommandRequest struct {
    Command []string `json:"command"`
}

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}
