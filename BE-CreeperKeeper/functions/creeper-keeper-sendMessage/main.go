package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

var (
	apiClient *apigatewaymanagementapi.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	apiClient = apigatewaymanagementapi.NewFromConfig(cfg)
}

// Define the message structure expected
type WebSocketMessage struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID
	var msg WebSocketMessage

	// Parse the incoming message
	if err := json.Unmarshal([]byte(event.Body), &msg); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	switch msg.Action {
	case "sendLog":
		// Send log data received from EC2 to the frontend
		err := sendMessageToClient(connectionID, msg.Data)
		if err != nil {
			log.Printf("Error sending message to connection: %v", err)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
		}
	default:
		log.Printf("Unsupported action: %s", msg.Action)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func sendMessageToClient(connectionID, message string) error {
	_, err := apiClient.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         []byte(message),
	})
	return err
}

func main() {
	lambda.Start(handler)
}
