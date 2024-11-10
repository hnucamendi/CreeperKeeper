package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hnucamendi/jwt-go/jwt"
)

var (
	j  *jwt.JWTClient
	sc *ssm.Client
)

func init() {
	// Loading default configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	sc = ssm.NewFromConfig(cfg)

	// Initializing JWT client
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

func getParams(ctx context.Context, paths ...string) (map[string]string, error) {
	params := map[string]string{}
	for _, path := range paths {
		param, err := sc.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           &path,
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, err
		}
		params[path] = *param.Parameter.Value
	}
	return params, nil
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// Checking for Authorization header
	authHeader, ok := event.Headers["Authorization"]
	if !ok || strings.TrimSpace(authHeader) == "" {
		log.Println("Authorization header missing")
		return generatePolicy("user", "Deny", "*"), nil
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	log.Printf("Received request for route: %s", event.RequestContext.RouteKey)

	// Retrieving SSM parameters
	p, err := getParams(ctx, "/accountID")
	if err != nil {
		log.Printf("Failed to get parameters: %v", err)
		return generatePolicy("user", "Deny", "*"), nil
	}

	// Constructing the resource ARN for the WebSocket API
	apiID := event.RequestContext.APIID
	stage := event.RequestContext.Stage
	region := "us-east-1" // This can be retrieved dynamically as well
	accountID := p["/accountID"]

	resourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/POST/%s",
		region, accountID, apiID, stage, event.RequestContext.RouteKey)

	// Validating the JWT token
	err = j.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return generatePolicy("user", "Deny", resourceArn), nil
	}

	log.Printf("Token validated successfully, generating allow policy")
	return generatePolicy("user", "Allow", resourceArn), nil
}

func main() {
	lambda.Start(handler)
}
