package docker

import (
    "fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
    "github.com/docker/docker/api/types/network"
    "github.com/docker/go-connections/nat"
    "os"
    "path/filepath"
)

// VPSContainer đại diện cho một VPS container
type VPSContainer struct {
    ID          string
    Name        string
    Image       string
    Ports       map[string]string // host:container
    Env         []string
    VolumePath  string
    CPUQuota    int64
    MemoryLimit int64
    Status      string
}

// CreateVPS tạo một container VPS mới
func (c *Client) CreateVPS(vps *VPSContainer) (string, error) {
    // Đảm bảo thư mục volume tồn tại
    if err := os.MkdirAll(vps.VolumePath, 0755); err != nil {
        return "", fmt.Errorf("không thể tạo thư mục volume: %w", err)
    }

    // Kéo image nếu chưa có
    if err := c.pullImage(vps.Image); err != nil {
        return "", fmt.Errorf("không thể kéo image: %w", err)
    }

    // Cấu hình port mapping
    portBindings := nat.PortMap{}
    exposedPorts := nat.PortSet{}
    
    for hostPort, containerPort := range vps.Ports {
        natPort := nat.Port(containerPort + "/tcp")
        portBindings[natPort] = []nat.PortBinding{
            {HostIP: "0.0.0.0", HostPort: hostPort},
        }
        exposedPorts[natPort] = struct{}{}
    }

    // Cấu hình volume mount
    mounts := []mount.Mount{
        {
            Type:   mount.TypeBind,
            Source: vps.VolumePath,
            Target: "/data",
        },
    }

    // Cấu hình container
    config := &container.Config{
        Image:        vps.Image,
        Env:          vps.Env,
        ExposedPorts: exposedPorts,
        Labels: map[string]string{
            "vps.manager": "true",
            "vps.id":      vps.ID,
            "vps.name":    vps.Name,
        },
    }

    // Cấu hình host
    hostConfig := &container.HostConfig{
        PortBindings: portBindings,
        Mounts:       mounts,
        Resources: container.Resources{
            CPUQuota:  vps.CPUQuota,
            Memory:    vps.MemoryLimit,
        },
        RestartPolicy: container.RestartPolicy{
            Name: "unless-stopped",
        },
    }

    // Cấu hình network
    networkConfig := &network.NetworkingConfig{}

    // Tạo container
    resp, err := c.cli.ContainerCreate(
        c.ctx,
        config,
        hostConfig,
        networkConfig,
        nil,
        vps.Name,
    )
    if err != nil {
        return "", fmt.Errorf("không thể tạo container: %w", err)
    }

    return resp.ID, nil
}

// StartVPS khởi động container
func (c *Client) StartVPS(containerID string) error {
    return c.cli.ContainerStart(c.ctx, containerID, types.ContainerStartOptions{})
}

// StopVPS dừng container
func (c *Client) StopVPS(containerID string) error {
    timeout := 10
    return c.cli.ContainerStop(c.ctx, containerID, container.StopOptions{
        Timeout: &timeout,
    })
}

// RemoveVPS xóa container và dữ liệu (tùy chọn)
func (c *Client) RemoveVPS(containerID string, removeVolumes bool) error {
    return c.cli.ContainerRemove(c.ctx, containerID, types.ContainerRemoveOptions{
        Force:         true,
        RemoveVolumes: removeVolumes,
    })
}

// GetVPSStatus lấy trạng thái của container
func (c *Client) GetVPSStatus(containerID string) (string, error) {
    inspect, err := c.cli.ContainerInspect(c.ctx, containerID)
    if err != nil {
        return "", err
    }
    return inspect.State.Status, nil
}

// ListVPS liệt kê tất cả container VPS
func (c *Client) ListVPS() ([]types.Container, error) {
    containers, err := c.cli.ContainerList(c.ctx, types.ContainerListOptions{
        All: true,
        Filters: filters.NewArgs(
            filters.Arg("label", "vps.manager=true"),
        ),
    })
    return containers, err
}

// ExecCommand thực thi lệnh trong container
func (c *Client) ExecCommand(containerID string, command []string) (string, error) {
    // Tạo exec configuration
    execConfig := types.ExecConfig{
        AttachStdout: true,
        AttachStderr: true,
        Cmd:          command,
    }

    // Tạo exec instance
    execResp, err := c.cli.ContainerExecCreate(c.ctx, containerID, execConfig)
    if err != nil {
        return "", err
    }

    // Attach vào exec
    attachResp, err := c.cli.ContainerExecAttach(c.ctx, execResp.ID, types.ExecStartCheck{})
    if err != nil {
        return "", err
    }
    defer attachResp.Close()

    // Đọc output
    var output bytes.Buffer
    _, err = stdcopy.StdCopy(&output, &output, attachResp.Reader)
    if err != nil {
        return "", err
    }

    return output.String(), nil
}

// pullImage kéo image từ registry
func (c *Client) pullImage(imageName string) error {
    reader, err := c.cli.ImagePull(c.ctx, imageName, types.ImagePullOptions{})
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // Đọc output để hoàn thành pull
    io.Copy(io.Discard, reader)
    return nil
}
