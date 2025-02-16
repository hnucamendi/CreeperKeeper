package ckec2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
