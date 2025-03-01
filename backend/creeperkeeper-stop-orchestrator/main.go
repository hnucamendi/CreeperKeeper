package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hnucamendi/jwt-go/jwt"
)

type sqsEvent struct{}

type Message struct {
	ServerStatus              string `json:"serverStatus"`
	MinecraftServerStatus     string `json:"minecraftServerStatus"`
	LastServerStatus          string `json:"lastServerStatus"`
	LastMinecraftServerStatus string `json:"lastMinecraftServerStatus"`
}

var (
	c          *Clients
	ssmClient  *ssm.Client
	j          *jwt.JWTClient
	httpClient *http.Client
	baseURL    string = "https://api.creeperkeeper.com"
)

type Clients struct {
	j          *jwt.JWTClient
	ssm        *ssm.Client
	httpClient *http.Client
}

// {"serverStatus":"test", minecraftServerStatus:"test"}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
	}

	c.httpClient = http.DefaultClient
	c.ssm = ssm.NewFromConfig(cfg)
	id, secret, audience, url, err := getParameters(context.TODO(), c)
	if err != nil {
		return
	}

	c.j = jwt.NewJWTClient(
		jwt.JWTClientID(*id),
		jwt.JWTClientSecret(*secret),
		jwt.JWTAudience(*audience),
		jwt.JWTGrantType("client_credentials"),
		jwt.JWTTenantURL(*url),
	)
}

func handler(ctx context.Context, event events.SQSEvent) (string, error) {
	count := len(event.Records)
	msgs := make([]*Message, count)

	for i := range event.Records {
		err := json.Unmarshal([]byte(event.Records[i].Body), &msgs[i])
		if err != nil {
			return "", err
		}
	}

	ok, err := pingServer(ctx)
	if err != nil {
		return "", err
	}

	if !ok {

	}

	// ping server
	// ping mc server
	// if both running then update DB
	// if server not running, update state, return
	// if server running but mc server is not then update state, return

	return "", nil
}

func pingServer(ctx context.Context, c *Clients, wantStatus string) (bool, error) {
	url := baseURL + "/server/ping"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	var status string
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(b, &status)
	if err != nil {
		return false, err
	}

	if status != wantStatus {
		return false, nil
	}

	ok, err := pingMCServer()
	if err != nil {
		return false, err
	}

	return true, nil
}

func pingMCServer() (bool, error) {

	return true, nil
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
	clientID, err := getParameter(ctx, "/creeperkeeper/jwt/client/id", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	clientSecret, err := getParameter(ctx, "/creeperkeeper/jwt/client/secret", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	audience, err := getParameter(ctx, "/creeperkeeper/jwt/client/audience", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tenantURL, err := getParameter(ctx, "/creeperkeeper/jwt/client/url", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return clientID, clientSecret, audience, tenantURL, nil
}

func main() {
	lambda.Start(handler)
}
