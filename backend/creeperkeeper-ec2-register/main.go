package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hnucamendi/jwt-go/jwt"
)

type Detail struct {
	State      string
	InstanceID string
}

const (
	baseURL string = "https://api.creeperkeeper.com"
)

var (
	ec2Client  *ec2.Client
	ssmClient  *ssm.Client
	httpClient *http.Client
)

type Clients struct {
	ec2Client  *ec2.Client
	ssmClient  *ssm.Client
	httpClient *http.Client
}

func handler(ctx context.Context, event events.CloudWatchEvent) (string, error) {
	var detail *Detail
	err := json.Unmarshal(event.Detail, &detail)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshall event details")
	}

	fmt.Printf("%+v", detail)

	c, err := initAWSClients(ctx)
	if err != nil {
		return "", err
	}

	switch detail.State {
	case "running":
		fmt.Println("TAMO made it here", detail.State)
		_, err := handleRunningState(ctx, detail, c)
		if err != nil {
			fmt.Println("TAMO ERR CHECK", err)
			return "", fmt.Errorf("failed to register server on state: %s error: %w", detail.State, err)
		}

		return "Success", nil

	case "stopping":
		handleStoppingState(ctx, detail, c)
	default:
		return "", fmt.Errorf("invalid event state: %v", detail.State)
	}

	return "", nil
}

func initAWSClients(ctx context.Context) (*Clients, error) {
	c := &Clients{}
	if ec2Client != nil {
		c.ec2Client = ec2Client
	}

	if ssmClient != nil {
		c.ssmClient = ssmClient
	}

	if httpClient != nil {
		c.httpClient = httpClient
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	c.ec2Client = ec2.NewFromConfig(cfg)
	c.ssmClient = ssm.NewFromConfig(cfg)
	c.httpClient = &http.Client{
		Timeout: 2 * time.Minute,
	}
	return c, nil
}

func handleRunningState(ctx context.Context, detail *Detail, clients *Clients) (bool, error) {
	ip, name, err := getInstanceDetails(ctx, &detail.InstanceID, clients.ec2Client)
	if err != nil {
		return false, err
	}

	fmt.Println("TAMO IP", ip)
	fmt.Println("TAMO NAME", name)

	clientID, err := getParameter(ctx, "/creeperkeeper/jwt/client/id", clients.ssmClient)
	if err != nil {
		return false, err
	}
	clientSecret, err := getParameter(ctx, "/creeperkeeper/jwt/client/secret", clients.ssmClient)
	if err != nil {
		return false, err
	}
	audience, err := getParameter(ctx, "/creeperkeeper/jwt/client/audience", clients.ssmClient)
	if err != nil {
		return false, err
	}

	jc := jwt.NewJWTClient(
		jwt.JWTClientID(*clientID),
		jwt.JWTClientSecret(*clientSecret),
		jwt.JWTAudience(*audience),
	)

	token, err := jc.GenerateToken(http.DefaultClient)
	if err != nil {
		return false, err
	}

	fmt.Println("TAMO", token)

	body := map[string]*string{
		"serverID":   &detail.InstanceID,
		"serverIP":   ip,
		"serverName": name,
	}

	jbody, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", baseURL+"/server/register", bytes.NewBuffer(jbody))
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := clients.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return false, fmt.Errorf("failed to register server in DB %v", res.Status)
	}

	return true, nil
}

func handleStoppingState(ctx context.Context, detail *Detail, clients *Clients) {
	// TODO: Implement Logic to save world data to S3
}

func getInstanceDetails(ctx context.Context, instanceID *string, ec *ec2.Client) (*string, *string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*instanceID},
	}

	out, err := ec.DescribeInstances(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return nil, nil, fmt.Errorf("instance not found")
	}

	ip := out.Reservations[0].Instances[0].PublicIpAddress
	if ip == nil {
		return nil, nil, fmt.Errorf("instance does not have a public IP address")
	}

	var name *string
	for i := 0; i < len(out.Reservations[0].Instances[0].Tags); i++ {
		if *out.Reservations[0].Instances[0].Tags[i].Key == "Name" {
			name = out.Reservations[0].Instances[0].Tags[i].Value
		}
	}

	return ip, name, nil
}

func getParameter(ctx context.Context, path string, ssmClient *ssm.Client) (*string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	out, err := ssmClient.GetParameter(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameter: %w", err)
	}
	return out.Parameter.Value, nil
}

func main() {
	lambda.Start(handler)
}
