package cpec2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2State int

const (
	STOP EC2State = iota
	START
	TERMINATE
)

func getInstanceIP(ctx context.Context, client *ec2.Client, instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	out, err := client.DescribeInstances(ctx, input)
	if err != nil {
		return "", err
	}

	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found")
	}

	ip := out.Reservations[0].Instances[0].PublicIpAddress
	if ip == nil {
		return "", fmt.Errorf("instance does not have a public IP address")
	}

	return *ip, nil
}

// - 0 : pending
// - 32 : shutting-down
//
// - 64 : stopping
// - 48 : terminated
//
// - 16 : running
// - 80 : stopped
func describeInstanceStatus(ctx context.Context, client *ec2.Client, instanceID string, desiredState EC2State) (string, error) {
	safeDescribeInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{instanceID},
		IncludeAllInstances: aws.Bool(true),
		DryRun:              aws.Bool(true), // Initial dry-run to check permissions
	}

	_, err := client.DescribeInstanceStatus(ctx, safeDescribeInput)
	if err != nil {
		// Check if the error is a DryRunOperation error
		if strings.Contains(err.Error(), "DryRunOperation") {
			// Dry run succeeded, proceed with actual request
			safeDescribeInput.DryRun = aws.Bool(false)
		} else {
			// If there's another error, return it
			return "", err
		}
	}

	out, err := client.DescribeInstanceStatus(ctx, safeDescribeInput)
	if err != nil {
		return "", err
	}

	// Check if InstanceStatuses is empty
	if len(out.InstanceStatuses) == 0 {
		return "", fmt.Errorf("instance status is not available for instance ID: %s", instanceID)
	}

	instanceStatus := out.InstanceStatuses[0]

	// Ensure that InstanceState is not nil
	if instanceStatus.InstanceState == nil {
		return "", fmt.Errorf("instance state information is not available for instance ID: %s", instanceID)
	}

	switch *instanceStatus.InstanceState.Code {
	case int32(0):
		return "", fmt.Errorf("instance is pending, try again later")
	case int32(16):
		if desiredState == START {
			ip, err := getInstanceIP(ctx, client, instanceID)
			if err != nil {
				return ip, fmt.Errorf("error getting instance IP address: %v", err)
			}
			return ip, nil
		}
	case int32(32):
		return "", fmt.Errorf("instance is being terminated, try again later")
	case int32(48):
		if desiredState == TERMINATE {
			return "", fmt.Errorf("instance is already terminated")
		}
	case int32(64):
		return "", fmt.Errorf("instance is stopping, try again later")
	case int32(80):
		if desiredState == STOP {
			return "", fmt.Errorf("instance is already stopped")
		}
	}

	return "", nil
}

func StartEC2Instance(ctx context.Context, client *ec2.Client, instanceID string) (map[string]string, error) {
	var ip string
	var successMessage string
	safeStartInput := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(true),
	}

	ip, err := describeInstanceStatus(ctx, client, instanceID, START)
	if err != nil {
		return nil, err
	}

	_, err = client.StartInstances(ctx, safeStartInput)
	if err != nil {
		if strings.Contains(err.Error(), "DryRunOperation") {
			safeStartInput.DryRun = aws.Bool(false)
			_, err := client.StartInstances(ctx, safeStartInput)
			if err != nil {
				return nil, fmt.Errorf("error starting instance: %v", err)
			}

			// Check if the instance is in the "pending" state
			successMessage = fmt.Sprintf("Instance %s is pending to start", instanceID)

			// Get the public IP address of the instance
			ip, err = getInstanceIP(ctx, client, instanceID)
			if err != nil {
				return map[string]string{"ip": ip, "success": successMessage}, fmt.Errorf("error getting instance IP address: %v", err)
			}

		} else {
			return map[string]string{"ip": ip, "success": successMessage}, fmt.Errorf("error starting instance: %v ip: %v", err, ip)
		}
	}
	return map[string]string{"ip": ip, "success": successMessage}, nil
}

func StopEC2Instance(ctx context.Context, client *ec2.Client, instanceID string) error {
	safeStopInput := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(true),
	}

	if _, err := describeInstanceStatus(ctx, client, instanceID, STOP); err != nil {
		return err
	}

	_, err := client.StopInstances(ctx, safeStopInput)
	if err != nil {
		if strings.Contains(err.Error(), "DryRunOperation") {
			safeStopInput.DryRun = aws.Bool(false)
			_, err = client.StopInstances(ctx, safeStopInput)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func TerminateEC2Instance(ctx context.Context, client *ec2.Client, db *dynamodb.Client, instanceID string) error {
	safeTerminateInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(true),
	}

	if _, err := describeInstanceStatus(ctx, client, instanceID, TERMINATE); err != nil {
		return err
	}

	_, err := client.TerminateInstances(ctx, safeTerminateInput)
	if err != nil {
		if strings.Contains(err.Error(), "DryRunOperation") {
			safeTerminateInput.DryRun = aws.Bool(false)
			_, err = client.TerminateInstances(ctx, safeTerminateInput)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Delete the item from the DynamoDB table
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("CreeperKeeper"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: instanceID},
			"SK": &types.AttributeValueMemberS{Value: "instance"},
		},
	}

	_, err = db.DeleteItem(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func WriteResponse(w http.ResponseWriter, code int, message interface{}) {
	w.WriteHeader(code)
	response := map[string]interface{}{"message": message}
	jMessage, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	w.Write(jMessage)
}
