package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func createResponse() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       "Testing Success",
	}
}

func HandleRequest(event events.APIGatewayV2HTTPRequest, ctx context.Context) {
	createResponse()
}

func main() {
	lambda.Start(HandleRequest)
}
