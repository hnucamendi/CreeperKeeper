package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/hnucamendi/jwt-go/jwt"
)

var (
	mux *http.ServeMux
	sc  *ssm.Client
	db  *dynamodb.Client
	j   *jwt.JWT
)

type C struct {
	sc *ssm.Client
	db *dynamodb.Client
	j  *jwt.JWT
	*http.Client
}

func init() {
	log.Println("Starting from Cold Start")
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

	j = &jwt.JWT{
		TenantURL: "https://dev-bxn245l6be2yzhil.us.auth0.com/oauth/token",
	}

	hc := &http.Client{}

	log.Printf("clients initiated: ssm: %v db: %v jwt: %v http: %v", sc != nil, db != nil, j != nil, hc != nil)

	c := &C{
		sc:     sc,
		db:     db,
		j:      j,
		Client: hc,
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
