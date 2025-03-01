package orchestrator

import (
	"context"

	"github.com/hnucamendi/creeper-keeper/service/orchestrator/sqs"
	"github.com/hnucamendi/creeper-keeper/types"
)

type Orchestrator interface {
	OrchestrateCallback(ctx context.Context, input types.OrchestratorMessage) error
}

type Client struct {
	*sqs.Client
}

func NewOrchestrator() *Client {
	c := &Client{}
	sqs, err := sqs.NewSQS()
	if err != nil {
		c.Client = nil
	}
	c.Client = sqs
	return c
}
