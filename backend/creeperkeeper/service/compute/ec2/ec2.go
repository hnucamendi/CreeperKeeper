package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hnucamendi/creeper-keeper/types"
)

type EC2API interface {
	DescribeInstanceStatus(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}

type Client struct {
	*ec2.Client
}

func (c *Client) GetServerStatus(ctx context.Context, serverID string) (*string, error) {
	state, err := getServerStatus(ctx, c.Client, serverID)
	if err != nil {
		return nil, err
	}
	status := ec2StateToString(state)
	return &status, nil
}

func (c *Client) StartServer(ctx context.Context, serverID string) error {
	status, err := getServerStatus(ctx, c.Client, serverID)
	if err != nil {
		return err
	}

	if status == types.STOPPING || status == types.TERMINATED || status == types.SHUTTINGDOWN || status == types.PENDING || status == types.NOTFOUND {
		return fmt.Errorf("EC2 is in an invalid state, code: %v", status)
	}

	if status == types.STOPPED {
		startInput := &ec2.StartInstancesInput{
			InstanceIds: []string{serverID},
		}
		_, err := c.Client.StartInstances(ctx, startInput)
		if err != nil {
			return fmt.Errorf("error starting instance: %v", err)
		}
	}

	return nil
}

func (c *Client) StopServer(ctx context.Context, serverID string) error {
	stopInput := &ec2.StopInstancesInput{
		InstanceIds: []string{serverID},
	}
	_, err := c.Client.StopInstances(ctx, stopInput)
	if err != nil {
		return err
	}
	return nil
}

// - 0 : pending
// - 32 : shutting-down
//
// - 64 : stopping
// - 48 : terminated
//
// - 16 : running
// - 80 : stopped
func getServerStatus(ctx context.Context, client *ec2.Client, serverID string) (types.EC2State, error) {
	describeInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{serverID},
		IncludeAllInstances: aws.Bool(true),
	}

	out, err := client.DescribeInstanceStatus(ctx, describeInput)
	if err != nil {
		return types.NOTFOUND, err
	}

	// Check if InstanceStatuses is empty
	if len(out.InstanceStatuses) == 0 {
		return types.NOTFOUND, fmt.Errorf("instance status is not available for instance ID: %s", serverID)
	}

	instanceStatus := *out.InstanceStatuses[0].InstanceState.Code

	switch instanceStatus {
	case 0:
		return types.PENDING, nil
	case 32:
		return types.SHUTTINGDOWN, nil
	case 64:
		return types.STOPPING, nil
	case 48:
		return types.TERMINATED, nil
	case 16:
		return types.RUNNING, nil
	case 80:
		return types.STOPPED, nil
	default:
		return types.NOTFOUND, nil
	}
}

func ec2StateToString(state types.EC2State) string {
	switch state {
	case 0:
		return "PENDING"
	case 1:
		return "SHUTTINGDOWN"
	case 2:
		return "STOPPING"
	case 3:
		return "TERMINATED"
	case 4:
		return "RUNNING"
	case 5:
		return "STOPPED"
	default:
		return "NOTFOUND"
	}
}

func NewCompute() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: ec2.NewFromConfig(cfg),
	}, nil
}
