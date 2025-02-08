package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var (
	mux *http.ServeMux
)

func init() {
	mux = http.NewServeMux()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	ec2 := ec2.NewFromConfig(cfg)
	db := dynamodb.NewFromConfig(cfg)

	c := &C{
		EC2: ec2,
		DB:  db,
	}

	h := NewHandler(c)
	loadRoutes(mux, h)

}

func handler(context context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return httpadapter.NewV2(mux).ProxyWithContext(context, event)
}

func main() {
	lambda.Start(handler)
}
