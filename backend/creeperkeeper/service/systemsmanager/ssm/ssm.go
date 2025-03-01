package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type SSMAPI interface {
	SendCommand(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error)
}

type Client struct {
	*ssm.Client
}

func (c *Client) Send(ctx context.Context, serverID string, commands []string) error {
	cmdInput := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{serverID},
		CloudWatchOutputConfig: &ssmTypes.CloudWatchOutputConfig{
			CloudWatchOutputEnabled: true,
			CloudWatchLogGroupName:  aws.String("/aws/lambda/creeperkeeper"),
		},
		Parameters: map[string][]string{
			"commands":         commands,
			"workingDirectory": {"/home/ec2-user"},
		},
	}
	_, err := c.Client.SendCommand(ctx, cmdInput)
	if err != nil {
		return err
	}
	return nil
}

func NewSSM() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: ssm.NewFromConfig(cfg),
	}, nil
}
