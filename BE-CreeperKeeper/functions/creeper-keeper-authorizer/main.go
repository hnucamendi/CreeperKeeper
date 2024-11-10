package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hnucamendi/CreeperKeeper/BE-CreeperKeeper/functions/creeper-keeper-authorizer/vendor/github.com/hnucamendi/jwt-go/jwt"
)

var (
	j *jwt.JWTClient
)

func init() {
	j = jwt.NewJWTClient(
		jwt.JWTTenantURL("https://dev-bxn245l6be2yzhil.us.auth0.com"),
	)
}

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
	token := strings.TrimSpace(strings.TrimPrefix(event.AuthorizationToken, "Bearer"))
	log.Printf("Received message %+v", event)

	err := j.ValidateToken(token)
	if err != nil {
		return generateDenyPolicy(event.MethodArn), nil
	}

	return generateAllowPolicy(event.MethodArn), nil
}

func main() {
	lambda.Start(handler)
}
