package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type AuthResponse struct {
	PrincipalId    string         `json:"principalId"`
	PolicyDocument PolicyDocument `json:"policyDocument,omitempty"`
	Context        Context        `json:"context"`
}

type Context struct {
	StringKey  string `json:"stringKey"`
	NumberKey  int    `json:"numberKey"`
	BooleanKey bool   `json:"booleanKey"`
}

type PolicyDocument struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Action   string `json:"Action"`
	Effect   string `json:"Effect"`
	Resource string `json:"Resource"`
}

func generateAllow(principalId string, resource string) *AuthResponse {
	return generatePolicy(principalId, "Allow", resource)
}

func generatePolicy(principalId string, effect string, resource string) *AuthResponse {
	policyDocument := PolicyDocument{
		Version: "2012-10-17",
		Statement: []Statement{
			{
				Action:   "execute-api:Invoke",
				Effect:   effect,
				Resource: resource,
			},
		},
	}

	authResponse := &AuthResponse{
		PrincipalId:    principalId,
		PolicyDocument: policyDocument,
		Context: Context{
			StringKey:  "stringval",
			NumberKey:  123,
			BooleanKey: true,
		},
	}

	return authResponse
}

func HandleRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (string, error) {
	authResponse := generateAllow("principalId_value", "resource_value")
	authResponseJSON, err := json.Marshal(authResponse)
	if err != nil {
		return "", err
	}

	return string(authResponseJSON), nil
}

func main() {
	lambda.Start(HandleRequest)
}
