package docker

import (
    "context"
    "github.com/docker/docker/client"
)

type Client struct {
    cli *client.Client
    ctx context.Context
}

func NewClient(host string) (*Client, error) {
    cli, err := client.NewClientWithOpts(
        client.WithHost(host),
        client.WithAPIVersionNegotiation(),
    )
    if err != nil {
        return nil, err
    }
    return &Client{
        cli: cli,
        ctx: context.Background(),
    }, nil
}

func (c *Client) Close() error {
    return c.cli.Close()
}

func (c *Client) Ping() error {
    _, err := c.cli.Ping(c.ctx)
    return err
}
