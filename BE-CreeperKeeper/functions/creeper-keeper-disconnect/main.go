package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) error {
	log.Printf("Received message %+v", event)

	return nil
}

func main() {
	lambda.Start(handler)
}
