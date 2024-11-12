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
	return events.APIGatewayCustomAuthorizerResponse{
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
}

func generateAllow(principalID, resource string) events.APIGatewayCustomAuthorizerResponse {
	return generatePolicy(principalID, "Allow", resource)
}

func generateDeny(principalID, resource string) events.APIGatewayCustomAuthorizerResponse {
	return generatePolicy(principalID, "Deny", resource)
}

func getParams(ctx context.Context, paths ...string) (map[string]string, error) {
	params := map[string]string{}

	result, err := sc.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	if len(result.InvalidParameters) > 0 {
		log.Printf("Invalid parameters: %v", result.InvalidParameters)
		return nil, fmt.Errorf("invalid parameters: %v", result.InvalidParameters)
	}

	for _, p := range result.Parameters {
		params[*p.Name] = *p.Value
	}
	return params, nil
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	apiID := event.RequestContext.APIID
	region := sc.Options().Region
	// Retrieving SSM parameters
	p, err := getParams(ctx, "/accountID")
	if err != nil {
		log.Printf("Failed to get parameters: %v", err)
		return generateDeny("user", fmt.Sprintf("arn:aws:execute-api:%s:1111111111:%s/*/*", region, apiID)), err
	}

	accountID, exists := p["/accountID"]
	if !exists {
		return generateDeny("user", fmt.Sprintf("arn:aws:execute-api:%s:1111111111:%s/*/*", region, apiID)), fmt.Errorf("/accountID not found")
	}

	resourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*", region, accountID, apiID)

	authHeader, ok := event.Headers["authorization"] // Check lowercase
	if !ok {
		authHeader, ok = event.Headers["Authorization"] // Fallback
	}
	if !ok || strings.TrimSpace(authHeader) == "" {
		log.Println("Authorization header missing or empty")
		return generateDeny("user", resourceArn), nil
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	log.Printf("Received request for route: %s", event.RequestContext.RouteKey)

	// Validating the JWT token
	err = j.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return generateDeny("user", resourceArn), nil
	}

	log.Printf("Token validated successfully, generating allow policy")
	return generateAllow("user", resourceArn), nil
}

func main() {
	lambda.Start(handler)
}
