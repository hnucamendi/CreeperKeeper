package ckec2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2State int

const (
	STOP EC2State = iota
	START
	TERMINATE
)

func getInstanceIP(ctx context.Context, client *ec2.Client, instanceID *string) (*string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceID},
	}

	out, err := client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found")
	}

	ip := out.Reservations[0].Instances[0].PublicIpAddress
	if ip == nil {
		return nil, fmt.Errorf("instance does not have a public IP address")
	}

	return ip, nil
}

// - 0 : pending
// - 32 : shutting-down
//
// - 64 : stopping
// - 48 : terminated
//
// - 16 : running
// - 80 : stopped
func describeInstanceStatus(ctx context.Context, client *ec2.Client, instanceID string, desiredState EC2State) (types.InstanceStatus, error) {
	describeInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{instanceID},
		IncludeAllInstances: aws.Bool(true),
	}

	out, err := client.DescribeInstanceStatus(ctx, describeInput)
	if err != nil {
		return types.InstanceStatus{}, err
	}

	// Check if InstanceStatuses is empty
	if len(out.InstanceStatuses) == 0 {
		return types.InstanceStatus{}, fmt.Errorf("instance status is not available for instance ID: %s", instanceID)
	}

	instanceStatus := out.InstanceStatuses[0]

	return instanceStatus, nil
}

func StartEC2Instance(ctx context.Context, client *ec2.Client, serverID *string) (*string, error) {
	startInput := &ec2.StartInstancesInput{
		InstanceIds: []string{*serverID},
	}
	out, err := client.StartInstances(ctx, startInput)
	if err != nil {
		return nil, fmt.Errorf("error starting instance: %v", err)
	}

	fmt.Printf("SERVER METADATA: %+v", out.ResultMetadata)

	// Get the public IP address of the instance
	ip, err := getInstanceIP(ctx, client, serverID)
	if err != nil {
		return nil, fmt.Errorf("error getting instance IP address: %v", err)
	}

	return ip, nil
}

func StopEC2Instance(ctx context.Context, client *ec2.Client, serverID *string) error {
	stopInput := &ec2.StopInstancesInput{
		InstanceIds: []string{*serverID},
	}
	out, err := client.StopInstances(ctx, stopInput)
	if err != nil {
		return err
	}

	fmt.Printf("Server META STOP %+v", out.ResultMetadata)

	// TODO: Make sure server is stopped before returning

	return nil
}

func Retry[T any](ctx context.Context, fn func() (T, error), attempts int) (T, error) {
	var zero T
	var result T
	var err error
	var maxDelay = 1 * time.Minute

	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Printf("attempt %d failed: %v", i+1, err)

			baseDelay := time.Duration(1<<i) * time.Second
			jitter := time.Duration(rand.Int63n(int64(baseDelay)))
			delay := baseDelay + jitter
			if delay > maxDelay {
				delay = maxDelay + time.Duration(rand.Int63n(int64(5*time.Second)))
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
		result, err = fn()
		if err == nil {
			return result, nil
		}
	}
	return zero, fmt.Errorf("after %d attempts, last error: %w", attempts, err)
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

// func TerminateEC2Instance(ctx context.Context, client *ec2.Client, db *dynamodb.Client, instanceID string) error {
// 	safeTerminateInput := &ec2.TerminateInstancesInput{
// 		InstanceIds: []string{instanceID},
// 		DryRun:      aws.Bool(true),
// 	}
//
// 	if _, err := describeInstanceStatus(ctx, client, instanceID, TERMINATE); err != nil {
// 		return err
// 	}
//
// 	_, err := client.TerminateInstances(ctx, safeTerminateInput)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "DryRunOperation") {
// 			safeTerminateInput.DryRun = aws.Bool(false)
// 			_, err = client.TerminateInstances(ctx, safeTerminateInput)
// 			if err != nil {
// 				return err
// 			}
// 		} else {
// 			return err
// 		}
// 	}
//
// 	// Delete the item from the DynamoDB table
// 	input := &dynamodb.DeleteItemInput{
// 		TableName: aws.String("CreeperKeeper"),
// 		Key: map[string]types.AttributeValue{
// 			"PK": &types.AttributeValueMemberS{Value: instanceID},
// 			"SK": &types.AttributeValueMemberS{Value: "instance"},
// 		},
// 	}
//
// 	_, err = db.DeleteItem(ctx, input)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
