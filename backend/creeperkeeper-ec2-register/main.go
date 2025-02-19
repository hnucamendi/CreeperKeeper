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
	State      string `json:"state"`
	InstanceID string `json:"instance-id"`
}

const (
	baseURL string = "https://api.creeperkeeper.com"
)

var (
	ec2Client  *ec2.Client
	ssmClient  *ssm.Client
	httpClient *http.Client
	jwtClient  *jwt.JWTClient
)

type Server struct {
	ID          *string `json:"serverID" dynamodbav:"PK"`
	SK          *string `json:"row" dynamodbav:"SK"`
	IP          *string `json:"serverIP" dynamodbav:"ServerIP"`
	Name        *string `json:"serverName" dynamodbav:"ServerName"`
	LastUpdated *string `json:"lastUpdated" dynamodbav:"LastUpdated"`
	IsRunning   *bool   `json:"isRunning" dynamodbav:"IsRunning"`
}

type Clients struct {
	ec2Client  *ec2.Client
	ssmClient  *ssm.Client
	httpClient *http.Client
	jwtClient  *jwt.JWTClient
}

func handler(ctx context.Context, event events.CloudWatchEvent) (string, error) {
	var detail *Detail
	err := json.Unmarshal(event.Detail, &detail)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshall event details")
	}

	c, err := initAWSClients(ctx)
	if err != nil {
		return "", err
	}

	switch detail.State {
	case "running":
		err := handleRunningState(ctx, detail, c)
		if err != nil {
			return "", fmt.Errorf("failed to register server on state: %q error: %w", detail.State, err)
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

	if jwtClient != nil {
		c.jwtClient = jwtClient
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

	id, secret, audience, url, err := getParameters(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to load parameters for JWT client")
	}

	c.jwtClient = jwt.NewJWTClient(
		jwt.JWTClientID(*id),
		jwt.JWTClientSecret(*secret),
		jwt.JWTAudience(*audience),
		jwt.JWTGrantType("client_credentials"),
		jwt.JWTTenantURL(*url),
	)
	return c, nil
}

func handleRunningState(ctx context.Context, detail *Detail, clients *Clients) error {
	ip, name, err := getInstanceDetails(ctx, &detail.InstanceID, clients.ec2Client)
	if err != nil {
		return err
	}

	_, err = clients.jwtClient.GenerateToken(clients.httpClient)
	if err != nil {
		return err
	}

	// TODO: wrap these two functions in go routines
	err = registerServerDetails(clients, &detail.InstanceID, ip, name)
	if err != nil {
		return fmt.Errorf("failed to register server %w", err)
	}

	err = startServer(ctx, clients, &detail.InstanceID, name)
	if err != nil {
		return fmt.Errorf("failed to start minecraft server %w", err)
	}

	return nil
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

func getParameters(ctx context.Context, c *Clients) (*string, *string, *string, *string, error) {
	clientID, err := getParameter(ctx, "/creeperkeeper/jwt/client/id", c.ssmClient)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	clientSecret, err := getParameter(ctx, "/creeperkeeper/jwt/client/secret", c.ssmClient)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	audience, err := getParameter(ctx, "/creeperkeeper/jwt/client/audience", c.ssmClient)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tenantURL, err := getParameter(ctx, "/creeperkeeper/jwt/client/url", c.ssmClient)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return clientID, clientSecret, audience, tenantURL, nil
}

func startServer(ctx context.Context, clients *Clients, serverID *string, serverName *string) error {
	cmds := []string{"sudo docker start " + *serverName}
	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{*serverID},
		Parameters: map[string][]string{
			"commands":         cmds,
			"workingDirectory": {"/home/ec2-user"},
		},
	}

	_, err := clients.ssmClient.SendCommand(ctx, input)
	if err != nil {
		return fmt.Errorf("ERROR TAMO %v", err)
	}

	return nil
}

func registerServerDetails(c *Clients, serverID *string, serverIP *string, serverName *string) error {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	lastUpdated := time.Now().In(zone).Format(time.DateTime)
	sk := "serverdetails"
	isRunning := true

	body := &Server{
		ID:          serverID,
		SK:          &sk,
		IP:          serverIP,
		Name:        serverName,
		LastUpdated: &lastUpdated,
		IsRunning:   &isRunning,
	}

	jbody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/server/register", bytes.NewBuffer(jbody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.jwtClient.AuthToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("failed to register server in DB %v", res.Status)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
