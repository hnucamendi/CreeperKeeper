package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hnucamendi/jwt-go/jwt"
)

var j *jwt.JWTClient

func init() {
	j = jwt.NewJWTClient(
		jwt.JWTTenantURL("https://dev-bxn245l6be2yzhil.us.auth0.com"),
	)
}

func generatePolicy(principalID, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	policy := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		},
	}
	return policy
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := strings.TrimSpace(strings.TrimPrefix(event.Headers["Authorization"], "Bearer "))
	log.Printf("Received message %+v", event)

	err := j.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return generatePolicy("user", "Deny", event.RequestContext.ResourcePath), nil
	}

	return generatePolicy("user", "Allow", event.RequestContext.ResourcePath), nil
}

func main() {
	lambda.Start(handler)
}
