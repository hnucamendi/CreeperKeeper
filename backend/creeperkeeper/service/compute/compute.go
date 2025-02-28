package compute

import (
	"context"

	"github.com/hnucamendi/creeper-keeper/service/compute/ec2"
)

type Compute interface {
	GetServerStatus(ctx context.Context, serverID string) (*string, error)
	StartServer(ctx context.Context, serverID string) error
	StopServer(ctx context.Context, serverID string) error
}

type Client struct {
	Client Compute
}

func NewCompute() *Client {
	c := &Client{}
	comp, err := ec2.NewCompute()
	if err != nil {
		c.Client = nil
	}
	c.Client = comp

	return c
}
