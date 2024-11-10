package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hnucamendi/jwt-go/jwt"
)

var (
	j  *jwt.JWTClient
	sc *ssm.Client
)

func init() {
	j = jwt.NewJWTClient(
		jwt.JWTTenantURL("https://dev-bxn245l6be2yzhil.us.auth0.com"),
	)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	ssm = ssm.NewFromConfig(cfg)
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

func getParams(paths ...string) (map[string]string, error) {
	params := map[string]string{}
	for _, path := range paths {
		param, err := sc.GetParameter(context.TODO(), &ssm.GetParameterInput{
			Name:           &path,
			WithDecryption: true,
		})
		if err != nil {
			log.Fatalf("failed to get parameter %s: %v", path, err)
		}
		params[path] = param.Parameter.Value
	}
	return params, nil
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := strings.TrimSpace(strings.TrimPrefix(event.Headers["Authorization"], "Bearer "))
	log.Printf("Received message %+v", event)

	p, err := getParams("/accountID")
	if err != nil {
		log.Fatalf("failed to get parameters: %v", err)
	}

	// Validate the JWT token here (assuming you have a function for that)
	err = j.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return generatePolicy("user", "Deny", "*"), nil
	}

	// Construct the resource ARN for the WebSocket API
	apiID := event.RequestContext.APIID
	stage := event.RequestContext.Stage
	region := "us-east-1"
	accountID := p["/accountID"]

	// resourceArn := "arn:aws:execute-api:" + region + ":" + accountID + ":" + apiID + "/" + stage + "/POST/*"
	resourceArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/POST/%s",
		region, accountID, apiID, stage, event.RequestContext.RouteKey)

	log.Printf("Token validated successfully, generating allow policy")
	return generatePolicy("user", "Allow", resourceArn), nil
}

func main() {
	lambda.Start(handler)
}
