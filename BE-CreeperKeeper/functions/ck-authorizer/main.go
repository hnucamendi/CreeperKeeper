package main

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func generatePolicy(principalID, effect, resource string) (events.APIGatewayCustomAuthorizerResponse, error) {
	if effect != "Allow" && effect != "Deny" {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("effect must be either 'Allow' or 'Deny'")
	}

	policyDocument := events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action:   []string{"execute-api:Invoke"},
				Effect:   effect,
				Resource: []string{resource},
			},
		},
	}

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID:    principalID,
		PolicyDocument: policyDocument,
	}, nil
}

func handleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := event.AuthorizationToken
	methodArn := event.MethodArn

	// Your custom authorization logic goes here.
	// For demonstration, we'll just check if the token is "allow" or "deny"
	switch strings.ToLower(token) {
	case "allow":
		return generatePolicy("user", "Allow", methodArn)
	case "deny":
		return generatePolicy("user", "Deny", methodArn)
	default:
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
}

func main() {
	lambda.Start(handleRequest)
}
