package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hnucamendi/jwt-go/jwt"
)

type Message struct {
	Server                   Server
	IsMinecraftServerRunning string `json:"isMinecraftServerRunning,omitempty"`
	IsServerRunning          string `json:"isServerRunning,omitempty"`
	Error                    string `json:"error,omitempty"`
}

type Server struct {
	ID          string `json:"serverID"`
	SK          string `json:"row"`
	IP          string `json:"serverIP"`
	Name        string `json:"serverName"`
	LastUpdated string `json:"lastUpdated"`
	IsRunning   bool   `json:"isRunning"`
}

type DLQMessage struct {
}

var (
	c          *Clients
	ssmClient  *ssm.Client
	sqsClient  *sqs.Client
	j          *jwt.JWTClient
	httpClient *http.Client
	dlqURL     *string
	baseURL    string = "https://api.creeperkeeper.com"
)

type STATUS string

const (
	RUNNING STATUS = "RUNNING"
	STOPPED STATUS = "STOPPED"
)

type Clients struct {
	j          *jwt.JWTClient
	ssm        *ssm.Client
	sqs        *sqs.Client
	httpClient *http.Client
}

// {"serverStatus":"test", minecraftServerStatus:"test"}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
	}

	c.httpClient = http.DefaultClient
	c.ssm = ssm.NewFromConfig(cfg)
	c.sqs = sqs.NewFromConfig(cfg)
	id, secret, audience, url, err := getParameters(context.TODO(), c)
	if err != nil {
		return
	}

	dlqURL, err = getParameter(context.TODO(), "/creeperkeeper/sqs/dlq", c.ssm)
	if err != nil {
		return
	}

	c.j = jwt.NewJWTClient(
		jwt.JWTClientID(*id),
		jwt.JWTClientSecret(*secret),
		jwt.JWTAudience(*audience),
		jwt.JWTGrantType("client_credentials"),
		jwt.JWTTenantURL(*url),
	)

}

func handler(ctx context.Context, event events.SQSEvent) (string, error) {
	count := len(event.Records)
	msgs := make([]*Server, count)

	for i := range event.Records {
		err := json.Unmarshal([]byte(event.Records[i].Body), &msgs[i])
		if err != nil {
			errors.Join(err)
		}

		serverRunning, err := pingServer(ctx, c, STOPPED)
		if err != nil {
			errors.Join(err)
		}

		mcServerRunning, err := pingMCServer(ctx, c, msgs[i].ID, msgs[i].Name)
		if err != nil {
			errors.Join(err)
		}

		if !serverRunning || !mcServerRunning {
			errorMessage := &Message{
				Server: Server{
					msgs[i].ID,
					msgs[i].SK,
					msgs[i].IP,
					msgs[i].Name,
					msgs[i].LastUpdated,
					msgs[i].IsRunning,
				},
				IsMinecraftServerRunning: strconv.FormatBool(mcServerRunning),
				IsServerRunning:          strconv.FormatBool(serverRunning),
				Error:                    err.Error(),
			}
			jvalue, _ := json.Marshal(errorMessage)
			dlqInput := &sqs.SendMessageInput{
				MessageBody: aws.String(string(jvalue)),
				QueueUrl:    aws.String(*dlqURL),
			}
			c.sqs.SendMessage(ctx, dlqInput)
		}

		err = registerServerDetails(c, msgs[i].ID, msgs[i].IP, msgs[i].Name)
		if err != nil {
			return "", errors.New(concat("failed to register server ", err.Error()))
		}
	}

	// ping server
	// ping mc server
	// if both running then update DB
	// if server not running, update state, return
	// if server running but mc server is not then update state, return

	return "", nil
}

func registerServerDetails(c *Clients, serverID string, serverIP string, serverName string) error {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	lastUpdated := time.Now().In(zone).Format(time.DateTime)
	sk := "serverdetails"
	isRunning := true

	body := &Server{
		ID:          serverID,
		SK:          sk,
		IP:          serverIP,
		Name:        serverName,
		LastUpdated: lastUpdated,
		IsRunning:   isRunning,
	}

	jbody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/server/register", bytes.NewBuffer(jbody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.j.AuthToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("failed to register server in DB " + res.Status)
	}

	return nil
}

func pingServer(ctx context.Context, c *Clients, wantStatus STATUS) (bool, error) {
	url := baseURL + "/server/ping"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	var status string
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(b, &status)
	if err != nil {
		return false, err
	}

	if status != string(wantStatus) {
		return false, nil
	}

	return true, nil
}

func pingMCServer(ctx context.Context, c *Clients, serverID string, serverName string) (bool, error) {
	cmds := []string{
		concat("echo $(docker ps -f \"publish=25565\" --format \"{{.ID}} {{.Names}}\" && docker inspect -f \"{{.State.Running}}\" ", serverName, ")"),
	}
	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{serverID},
		Parameters: map[string][]string{
			"commands":         cmds,
			"workingDirectory": {"/home/ec2-user"},
		},
	}

	out, err := c.ssm.SendCommand(ctx, input)
	if err != nil {
		return false, errors.New("failed to send command to EC2")
	}

	getInput := &ssm.GetCommandInvocationInput{
		CommandId:  out.Command.CommandId,
		InstanceId: aws.String(serverID),
	}
	cmdOut, err := c.ssm.GetCommandInvocation(ctx, getInput)
	if err != nil {
		return false, errors.New(concat("failed to get the command output for err", *out.Command.CommandId))
	}

	if cmdOut.StandardOutputContent == nil {
		return false, errors.New(*cmdOut.StandardErrorContent)
	}

	outputString := strings.Split(*cmdOut.StandardOutputContent, " ")
	fmt.Println(outputString)
	dockerContainerID := outputString[0]
	serviceName := outputString[1]
	isRunning := outputString[2]

	if (dockerContainerID == "" && serviceName != serverName) && isRunning == "false" {
		return false, nil
	}

	return true, nil
}

func getParameter(ctx context.Context, path string, ssmClient *ssm.Client) (*string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	out, err := ssmClient.GetParameter(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameter: %w", err)
	}
	return out.Parameter.Value, nil
}

func getParameters(ctx context.Context, c *Clients) (*string, *string, *string, *string, error) {
	clientID, err := getParameter(ctx, "/creeperkeeper/jwt/client/id", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	clientSecret, err := getParameter(ctx, "/creeperkeeper/jwt/client/secret", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	audience, err := getParameter(ctx, "/creeperkeeper/jwt/client/audience", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tenantURL, err := getParameter(ctx, "/creeperkeeper/jwt/client/url", c.ssm)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return clientID, clientSecret, audience, tenantURL, nil
}

func concat(strs ...string) string { return strings.Join(strs, "") }

func main() {
	lambda.Start(handler)
}
