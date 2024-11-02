package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var (
	mux *http.ServeMux
	sc  *ssm.Client
	db  *dynamodb.Client
)

type C struct {
	sc *ssm.Client
	db *dynamodb.Client
}

func init() {
	mux = http.NewServeMux()

	ssmcfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	dbcfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	sc = ssm.NewFromConfig(ssmcfg)
	db = dynamodb.NewFromConfig(dbcfg)

	c := &C{
		sc: sc,
		db: db,
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
