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
)

type EC2State int

const (
	PENDING EC2State = iota
	SHUTTINGDOWN
	STOPPING
	TERMINATED
	RUNNING
	STOPPED
	NOTFOUND
)

func getInstanceIP(ctx context.Context, client *ec2.Client, serverID *string) (*string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*serverID},
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
func getServerStatus(ctx context.Context, client *ec2.Client, serverID *string) (EC2State, error) {
	describeInput := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{*serverID},
		IncludeAllInstances: aws.Bool(true),
	}

	out, err := client.DescribeInstanceStatus(ctx, describeInput)
	if err != nil {
		return NOTFOUND, err
	}

	// Check if InstanceStatuses is empty
	if len(out.InstanceStatuses) == 0 {
		return NOTFOUND, fmt.Errorf("instance status is not available for instance ID: %s", *serverID)
	}

	instanceStatus := *out.InstanceStatuses[0].InstanceState.Code

	switch instanceStatus {
	case 0:
		return PENDING, nil
	case 32:
		return SHUTTINGDOWN, nil
	case 64:
		return STOPPING, nil
	case 48:
		return TERMINATED, nil
	case 16:
		return RUNNING, nil
	case 80:
		return STOPPED, nil
	default:
		return NOTFOUND, nil
	}
}

func StartEC2Instance(ctx context.Context, client *ec2.Client, serverID *string) error {
	if serverID == nil {
		return fmt.Errorf("serverID must not be nil: %v", serverID)
	}

	status, err := getServerStatus(ctx, client, serverID)
	if err != nil {
		return err
	}

	if status == STOPPING || status == TERMINATED || status == SHUTTINGDOWN || status == PENDING || status == NOTFOUND {
		return fmt.Errorf("EC2 is in an invalid state, code: %v", status)
	}

	if status == STOPPED {
		startInput := &ec2.StartInstancesInput{
			InstanceIds: []string{*serverID},
		}
		_, err := client.StartInstances(ctx, startInput)
		if err != nil {
			return fmt.Errorf("error starting instance: %v", err)
		}
	}

	return nil
}

func StopEC2Instance(ctx context.Context, client *ec2.Client, serverID *string) error {
	stopInput := &ec2.StopInstancesInput{
		InstanceIds: []string{*serverID},
	}
	_, err := client.StopInstances(ctx, stopInput)
	if err != nil {
		return err
	}
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
