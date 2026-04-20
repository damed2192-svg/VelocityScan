package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/yourusername/vps-manager/internal/config"
    "github.com/yourusername/vps-manager/internal/docker"
    "github.com/yourusername/vps-manager/internal/models"
)

type Handler struct {
    dockerClient *docker.Client
    config       *config.Config
    vpsStore     map[string]*models.VPS // Trong thực tế nên dùng database
}

func NewHandler(dockerClient *docker.Client, cfg *config.Config) *Handler {
    return &Handler{
        dockerClient: dockerClient,
        config:       cfg,
        vpsStore:     make(map[string]*models.VPS),
    }
}

// CreateVPS tạo một VPS mới
func (h *Handler) CreateVPS(w http.ResponseWriter, r *http.Request) {
    var req models.CreateVPSRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.jsonResponse(w, http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
        })
        return
    }

    // Set defaults
    if req.Image == "" {
        req.Image = h.config.DefaultImage
    }
    if req.Ports == nil {
        req.Ports = make(map[string]string)
    }

    // Tạo VPS record
    vpsID := uuid.New().String()
    vps := &models.VPS{
        ID:          vpsID,
        Name:        req.Name,
        Image:       req.Image,
        Ports:       req.Ports,
        Env:         req.Env,
        CPUQuota:    req.CPUQuota,
        MemoryLimit: req.MemoryLimit,
        CreatedAt:   time.Now(),
        Status:      "creating",
    }

    // Tạo container
    vpsContainer := &docker.VPSContainer{
        ID:          vpsID,
        Name:        fmt.Sprintf("vps-%s", req.Name),
        Image:       req.Image,
        Ports:       req.Ports,
        Env:         req.Env,
        VolumePath:  h.config.GetVolumePath(vpsID),
        CPUQuota:    req.CPUQuota,
        MemoryLimit: req.MemoryLimit,
    }

    containerID, err := h.dockerClient.CreateVPS(vpsContainer)
    if err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to create container: %v", err),
        })
        return
    }

    vps.ContainerID = containerID
    vps.Status = "created"
    h.vpsStore[vpsID] = vps

    h.jsonResponse(w, http.StatusCreated, models.APIResponse{
        Success: true,
        Message: "VPS created successfully",
        Data:    vps,
    })
}

// StartVPS khởi động VPS
func (h *Handler) StartVPS(w http.ResponseWriter, r *http.Request) {
    vpsID := chi.URLParam(r, "id")
    
    vps, exists := h.vpsStore[vpsID]
    if !exists {
        h.jsonResponse(w, http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "VPS not found",
        })
        return
    }

    if err := h.dockerClient.StartVPS(vps.ContainerID); err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to start VPS: %v", err),
        })
        return
    }

    vps.Status = "running"
    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Message: "VPS started",
        Data:    vps,
    })
}

// StopVPS dừng VPS
func (h *Handler) StopVPS(w http.ResponseWriter, r *http.Request) {
    vpsID := chi.URLParam(r, "id")
    
    vps, exists := h.vpsStore[vpsID]
    if !exists {
        h.jsonResponse(w, http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "VPS not found",
        })
        return
    }

    if err := h.dockerClient.StopVPS(vps.ContainerID); err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to stop VPS: %v", err),
        })
        return
    }

    vps.Status = "stopped"
    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Message: "VPS stopped",
        Data:    vps,
    })
}

// DeleteVPS xóa VPS
func (h *Handler) DeleteVPS(w http.ResponseWriter, r *http.Request) {
    vpsID := chi.URLParam(r, "id")
    
    vps, exists := h.vpsStore[vpsID]
    if !exists {
        h.jsonResponse(w, http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "VPS not found",
        })
        return
    }

    // Xóa container (có thể giữ lại volumes)
    removeVolumes := r.URL.Query().Get("remove_volumes") == "true"
    if err := h.dockerClient.RemoveVPS(vps.ContainerID, removeVolumes); err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to delete VPS: %v", err),
        })
        return
    }

    delete(h.vpsStore, vpsID)
    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Message: "VPS deleted",
    })
}

// ListVPS liệt kê tất cả VPS
func (h *Handler) ListVPS(w http.ResponseWriter, r *http.Request) {
    containers, err := h.dockerClient.ListVPS()
    if err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to list VPS: %v", err),
        })
        return
    }

    vpsList := make([]map[string]interface{}, 0)
    for _, container := range containers {
        vpsInfo := map[string]interface{}{
            "id":      container.ID,
            "name":    container.Names[0],
            "image":   container.Image,
            "status":  container.Status,
            "created": time.Unix(container.Created, 0),
            "ports":   container.Ports,
        }
        vpsList = append(vpsList, vpsInfo)
    }

    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Data:    vpsList,
    })
}

// GetVPS lấy thông tin chi tiết một VPS
func (h *Handler) GetVPS(w http.ResponseWriter, r *http.Request) {
    vpsID := chi.URLParam(r, "id")
    
    vps, exists := h.vpsStore[vpsID]
    if !exists {
        h.jsonResponse(w, http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "VPS not found",
        })
        return
    }

    // Cập nhật status từ Docker
    status, err := h.dockerClient.GetVPSStatus(vps.ContainerID)
    if err == nil {
        vps.Status = status
    }

    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Data:    vps,
    })
}

// ExecCommand thực thi lệnh trong VPS
func (h *Handler) ExecCommand(w http.ResponseWriter, r *http.Request) {
    vpsID := chi.URLParam(r, "id")
    
    var req models.ExecCommandRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.jsonResponse(w, http.StatusBadRequest, models.APIResponse{
            Success: false,
            Message: "Invalid request body",
        })
        return
    }

    vps, exists := h.vpsStore[vpsID]
    if !exists {
        h.jsonResponse(w, http.StatusNotFound, models.APIResponse{
            Success: false,
            Message: "VPS not found",
        })
        return
    }

    output, err := h.dockerClient.ExecCommand(vps.ContainerID, req.Command)
    if err != nil {
        h.jsonResponse(w, http.StatusInternalServerError, models.APIResponse{
            Success: false,
            Message: fmt.Sprintf("Failed to execute command: %v", err),
        })
        return
    }

    h.jsonResponse(w, http.StatusOK, models.APIResponse{
        Success: true,
        Data: map[string]string{
            "output": output,
        },
    })
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
