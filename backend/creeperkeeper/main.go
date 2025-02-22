package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/hnucamendi/jwt-go/jwt"
	"golang.org/x/exp/rand"
)

const (
	tableName string = "creeperkeeper"
)

var (
	mux *http.ServeMux
	sc  *ssm.Client
	db  *dynamodb.Client
	j   *jwt.JWT
	ec  *ec2.Client
)

type C struct {
	sc *ssm.Client
	db *dynamodb.Client
	j  *jwt.JWT
	ec *ec2.Client
	*http.Client
}

func init() {
	log.Println("Starting from Cold Start")
	mux = http.NewServeMux()
	rand.Seed(uint64(time.Now().UnixNano()))

	awscfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load sdk config" + err.Error())
	}

	sc = ssm.NewFromConfig(awscfg)
	db = dynamodb.NewFromConfig(awscfg)
	ec = ec2.NewFromConfig(awscfg)

	j = &jwt.JWT{
		TenantURL: "https://dev-bxn245l6be2yzhil.us.auth0.com/oauth/token",
	}

	hc := &http.Client{}

	c := &C{
		sc:     sc,
		db:     db,
		j:      j,
		ec:     ec,
		Client: hc,
	}

	h := NewHandler(c)
	loadRoutes(mux, h)

}

func handler(context context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	test, err := httpadapter.NewV2(mux).ProxyWithContext(context, event)
	fmt.Printf("%+v", test)
	return test, err
}

func main() {
	lambda.Start(handler)
}
