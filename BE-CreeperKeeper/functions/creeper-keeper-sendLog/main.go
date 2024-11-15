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

func init() {
	fmt.Println("Starting from Cold Start")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic(fmt.Sprintf("Unable to load SDK config, %v", err))
	}

	// Create an SSM client to fetch parameters
	ssmClient := ssm.NewFromConfig(cfg)

	// Get the APIID from SSM
	apiID, err := getParameter(ssmClient, "/ck/ws/APIID")
	if err != nil {
		panic(fmt.Sprintf("Unable to retrieve APIID from SSM, %v", err))
	}

	// Create the base URL for the API Gateway Management API client
	webSocketEndpoint := fmt.Sprintf("wss://%s.execute-api.us-east-1.amazonaws.com/ck/", apiID)

	// Create a custom API Gateway Management API client with the correct endpoint
	apiClient = apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		Region:           "us-east-1",
		EndpointResolver: apigatewaymanagementapi.EndpointResolverFromURL(webSocketEndpoint),
		Credentials:      cfg.Credentials,
	})
}

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

// WebSocketMessage represents the structure of messages expected from the WebSocket
type WebSocketMessage struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID
	log.Printf("Received message %+v", event)
	var msg WebSocketMessage

	// Parse the incoming message
	if err := json.Unmarshal([]byte(event.Body), &msg); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: "Invalid message format"}, nil
	}

	// Handle different actions based on the WebSocket message
	fmt.Printf("Message received: %+v\n", msg)

	// Handle sending log data to connected client
	err := sendMessageToClient(ctx, connectionID, msg.Data)
	if err != nil {
		if apiErr, ok := err.(*types.GoneException); ok {
			// Connection is no longer available (client disconnected)
			log.Printf("Connection %s is gone: %v", connectionID, apiErr)
		} else {
			log.Printf("Error sending message to connection %s: %v", connectionID, err)
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

	_, err := apiClient.PostToConnection(ctx, input)
	if err != nil {
		log.Printf("Error posting to connection %s: %v", connectionID, err)
	}
	return err
}

func main() {
	lambda.Start(handler)
}
