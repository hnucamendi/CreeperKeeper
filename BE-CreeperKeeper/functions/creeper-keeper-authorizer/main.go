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
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
	)
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

func generateAllow(principalID, resource string) events.APIGatewayCustomAuthorizerResponse {
	return generatePolicy(principalID, "Allow", resource)
}

func generateDeny(principalID, resource string) events.APIGatewayCustomAuthorizerResponse {
	return generatePolicy(principalID, "Deny", resource)
}

func getParams(ctx context.Context, paths ...string) (string, map[string]string, error) {
	params := map[string]string{}

	param, err := sc.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", nil, err
	}

	var t string
	for _, p := range param.Parameters {
		params[*p.Name] = *p.Value
		t = *p.Name
	}
	return t, params, nil
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// Retrieving SSM parameters
	n, p, err := getParams(ctx, "/accountID")
	if err != nil {
		log.Printf("Failed to get parameters: %v", err)
		return generateDeny("user", "*"), err
	}

	// Constructing the resource ARN for the WebSocket API
	apiID := event.RequestContext.APIID
	stage := event.RequestContext.Stage
	region := sc.Options().Region
	accountID := p[n]

	resourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/$connect",
		region, accountID, apiID, stage)

	authHeader, ok := event.Headers["authorization"] // Check lowercase
	if !ok {
		authHeader, ok = event.Headers["Authorization"] // Fallback
	}
	if !ok || strings.TrimSpace(authHeader) == "" {
		log.Println("Authorization header missing or empty")
		return generateDeny("me", resourceArn), err
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	log.Printf("Received request for route: %s", event.RequestContext.RouteKey)

	// Validating the JWT token
	err = j.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return generateDeny("user", resourceArn), err
	}

	log.Printf("Token validated successfully, generating allow policy")
	return generateAllow("user", resourceArn), nil
}

func main() {
	lambda.Start(handler)
}
