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

	// Initialize API Gateway Management client
	apiClient = apigatewaymanagementapi.NewFromConfig(cfg)

	// Base URL for sending messages to WebSocket connections
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
	switch msg.Action {
	case "sendLog":
		fmt.Printf("TAMO %+v\n", msg)
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
	default:
		// Unsupported action
		log.Printf("Unsupported action: %s", msg.Action)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: "Unsupported action"}, nil
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

	out, err := apiClient.PostToConnection(ctx, input)
	if err != nil {
		log.Printf("Error posting to connection %s: %v", connectionID, err)
	}
	fmt.Printf("OUT: %+v\n", out)
	return err
}

func main() {
	lambda.Start(handler)
}
