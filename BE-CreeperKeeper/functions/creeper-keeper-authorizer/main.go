package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hnucamendi/jwt-go/jwt"
)

func generateAllowPolicy(arn string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{arn},
				},
			},
		},
	}
}

func generateDenyPolicy(arn string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Deny",
					Resource: []string{arn},
				},
			},
		},
	}
}

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	j := jwt.NewJWTClient()

	log.Printf("Received message %+v", event)

	err := j.ValidateToken(event.AuthorizationToken)
	if err != nil {
		return generateAllowPolicy(event.MethodArn), nil
	}

	return generateAllowPolicy(event.MethodArn), nil
}

func main() {
	lambda.Start(handler)
}
