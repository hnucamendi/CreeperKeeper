package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/hnucamendi/creeper-keeper/service/compute"
	"github.com/hnucamendi/creeper-keeper/service/database"
	"github.com/hnucamendi/creeper-keeper/service/orchestrator"
	"github.com/hnucamendi/creeper-keeper/service/systemsmanager"
	"github.com/hnucamendi/jwt-go/jwt"
	"golang.org/x/exp/rand"
)

const (
	tableName string = "creeperkeeper"
)

var (
	dbClient             *database.Client
	computeClient        *compute.Client
	systemsmanagerClient *systemsmanager.Client
	orchestratorClient   *orchestrator.Client
	mux                  *http.ServeMux
	j                    *jwt.JWT
)

type C struct {
	db             *database.Client
	compute        *compute.Client
	systemsmanager *systemsmanager.Client
	orchestrator   *orchestrator.Client
	j              *jwt.JWT
	*http.Client
}

func init() {
	log.Println("Starting from Cold Start")
	mux = http.NewServeMux()
	rand.Seed(uint64(time.Now().UnixNano()))

	systemsmanagerClient = systemsmanager.NewSystemsManager()
	computeClient = compute.NewCompute()
	dbClient = database.NewDatabase(
		database.WithClient(database.DYNAMODB),
		database.WithTable(tableName),
	)

	j = &jwt.JWT{
		TenantURL: "https://dev-bxn245l6be2yzhil.us.auth0.com/oauth/token",
	}

	hc := &http.Client{}

	c := &C{
		db:             dbClient,
		compute:        computeClient,
		systemsmanager: systemsmanagerClient,
		orchestrator:   orchestratorClient,
		j:              j,
		Client:         hc,
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
