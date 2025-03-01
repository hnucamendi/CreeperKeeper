package sqs

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hnucamendi/creeper-keeper/types"
)

type SQSAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type Client struct {
	Client *sqs.Client
}

func (c *Client) OrchestrateCallback(ctx context.Context, inputBody types.OrchestratorMessage) error {
	bodyJson, err := json.Marshal(inputBody)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyJson)),
		QueueUrl:    aws.String(""),
	}
	out, err := c.Client.SendMessage(ctx, input)
	if err != nil {
		return err
	}

	if out.MD5OfMessageBody == nil {
		return errors.New("failed to send message to Queue")
	}

	return nil
}

func NewSQS() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: sqs.NewFromConfig(cfg),
	}, nil
}
