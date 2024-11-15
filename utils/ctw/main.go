package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hnucamendi/jwt-go/jwt"
	"golang.org/x/net/websocket"
)

const (
	LOG_FILE_PATH = "/home/ec2-user/Minecraft/logs/latest.log"
)

var (
	c *Client
)

type Client struct {
	*http.Client
	jwtClient *jwt.JWTClient
	ssmClient *ssm.Client
}

func initiate() {
	c = &Client{}

	c.Client = http.DefaultClient

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	c.ssmClient = ssm.NewFromConfig(cfg)

	p, err := getParams(context.TODO(),
		"/ck/jwt/client_id",
		"/ck/jwt/client_secret",
		"/ck/jwt/audience")
	if err != nil {
		log.Fatalf("Error getting parameters: %v", err)
	}

	clientID, ok := p["/ck/jwt/client_id"]
	if !ok {
		log.Fatalf("client_id not found in parameters")
	}

	clientSecret, ok := p["/ck/jwt/client_secret"]
	if !ok {
		log.Fatalf("client_secret not found in parameters")
	}

	audience, ok := p["/ck/jwt/audience"]
	if !ok {
		log.Fatalf("audience not found in parameters")
	}

	c.jwtClient = jwt.NewJWTClient(
		jwt.JWTClientID(clientID),
		jwt.JWTClientSecret(clientSecret),
		jwt.JWTAudience(audience),
		jwt.JWTGrantType("client_credentials"),
		jwt.JWTTenantURL("https://dev-bxn245l6be2yzhil.us.auth0.com"),
	)
}

func getToken() (string, error) {
	token, err := c.jwtClient.GenerateToken(c.Client)
	if err != nil {
		return "", err
	}

	return token, nil
}

func getParams(ctx context.Context, paths ...string) (map[string]string, error) {
	params := map[string]string{}

	result, err := c.ssmClient.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	if len(result.InvalidParameters) > 0 {
		log.Printf("Invalid parameters: %v", result.InvalidParameters)
		return nil, fmt.Errorf("invalid parameters: %v", result.InvalidParameters)
	}

	for _, p := range result.Parameters {
		params[*p.Name] = *p.Value
	}
	return params, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <log_message>", os.Args[0])
	}

	initiate()
	ctx := context.Background()

	token, err := getToken()
	if err != nil {
		log.Fatalf("Error getting token: %v", err)
	}

	p, err := getParams(ctx, "/ck/ws/APIID")
	if err != nil {
		log.Fatalf("Error getting parameters: %v", err)
	}

	APIID, ok := p["/ck/ws/APIID"]
	if !ok {
		log.Fatalf("APIID not found in parameters")
	}

	url := "wss://" + APIID + ".execute-api.us-east-1.amazonaws.com/ck/"

	// Create a new WebSocket configuration
	config, err := websocket.NewConfig(url, url)
	if err != nil {
		log.Fatalf("Error creating WebSocket config: %v", err)
	}

	// config.Header = http.Header{}
	config.Header.Set("Authorization", "Bearer "+token)

	// Connect to the WebSocket server
	conn, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatalf("Error connecting to WebSocket server: %v", err)
	}
	defer conn.Close()

	logMessage := os.Args[1]

	bodyMessage := map[string]string{
		"action": "sendLog",
		"data":   logMessage,
	}

	jBody, err := json.Marshal(bodyMessage)
	if err != nil {
		log.Fatalf("Error marshalling message: %v", err)
	}

	_, err = conn.Write(jBody)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
}
