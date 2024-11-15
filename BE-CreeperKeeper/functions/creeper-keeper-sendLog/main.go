package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var (
	apiClient *apigatewaymanagementapi.Client
)

// getParameter retrieves a parameter from AWS Systems Manager Parameter Store
func getParameter(ssmClient *ssm.Client, name string) (string, error) {
	param, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return *param.Parameter.Value, nil
}

func init() {
	fmt.Println("Starting from Cold Start")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic(fmt.Sprintf("Unable to load SDK config, %v", err))
	}

	// Create SSM client
	ssmClient := ssm.NewFromConfig(cfg)

	// Get API ID from SSM Parameter Store
	apiID, err := getParameter(ssmClient, "/ck/ws/APIID")
	if err != nil {
		panic(fmt.Sprintf("Unable to retrieve APIID from SSM, %v", err))
	}

	// Initialize API Gateway Management client
	apiClient = apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		Credentials:  cfg.Credentials,
		Region:       cfg.Region,
		BaseEndpoint: aws.String("https://" + apiID + ".execute-api.us-east-1.amazonaws.com/ck/@connections"), // Base URL for sending messages to WebSocket connections
	})

	// Base URL for sending messages to WebSocket connections
}

// WebSocketMessage represents the structure of messages expected from the WebSocket
type WebSocketMessage struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID
	log.Printf("Received event %+v\n", event)
	var msg WebSocketMessage

	// Parse the incoming message
	if err := json.Unmarshal([]byte(event.Body), &msg); err != nil {
		log.Printf("Error unmarshalling message: %v\n", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: "Invalid message format"}, nil
	}

	// Handle different actions based on the WebSocket message

	fmt.Printf("TAMO %+v\n", msg)
	// Handle sending log data to connected client
	err := sendMessageToClient(ctx, connectionID, msg.Data)
	if err != nil {
		if apiErr, ok := err.(*types.GoneException); ok {
			// Connection is no longer available (client disconnected)
			log.Printf("Connection %s is gone: %v\n", connectionID, apiErr)
		} else {
			log.Printf("Error sending message to connection %s: %v\n", connectionID, err)
		}
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "Failed to send message"}, nil
	}

	// Successfully processed the request
	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: "Message sent successfully"}, nil
}

// sendMessageToClient sends a message to a connected WebSocket client
func sendMessageToClient(ctx context.Context, connectionID, message string) error {
	input := &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         []byte(message),
	}

	fmt.Printf("INPUT: %+v\n", input)
	fmt.Printf("API CLIENT: %+v\n", apiClient)

	out, err := apiClient.PostToConnection(ctx, input)
	if err != nil {
		log.Printf("Error posting to connection %s: %v\n", connectionID, err)
	}
	fmt.Printf("OUT: %+v\n", out)
	return err
}

func main() {
	lambda.Start(handler)
}
