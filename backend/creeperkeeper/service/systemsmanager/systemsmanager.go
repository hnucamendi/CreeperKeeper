package systemsmanager

import (
	"context"

	"github.com/hnucamendi/creeper-keeper/service/systemsmanager/ssm"
)

type SystemsManager interface {
	Send(ctx context.Context, serverID string, serverName string) error
}

type Client struct {
	Client SystemsManager
}

func NewSystemsManager() *Client {
	c := &Client{}
	sysman, err := ssm.NewSSM()
	if err != nil {
		c.Client = nil
	}

	c.Client = sysman
	return c
}
