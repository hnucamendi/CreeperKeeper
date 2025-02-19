package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hnucamendi/creeper-keeper/ckec2"
)

type Handler struct {
	Client *C
}

func NewHandler(c *C) *Handler {
	return &Handler{
		Client: c,
	}
}

type Server struct {
	ID   *string `json:"serverID"`
	IP   *string `json:"serverIP"`
	Name *string `json:"serverName"`
}

func (ck *Server) unmarshallRequest(b io.ReadCloser) error {
	err := json.NewDecoder(b).Decode(&ck)
	if err != nil {
		return err
	}

	return nil
}

// Adds EC2 instance details to DynamoDB to be used by EC2 Directly
// TODO: take measures to ensure this cannot be invoked from FE
func (h *Handler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, http.StatusBadRequest, "serverID required for registering new server")
		return
	}

	if ck.IP == nil {
		WriteResponse(w, http.StatusBadRequest, "IP required for registering new server")
	}

	if ck.Name == nil {
		WriteResponse(w, http.StatusBadRequest, "server name is required for registering new server")
	}

	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, "failed to load timezone")
	}

	// TODO: Abstract DB logic in DB specific controller
	_, err = h.Client.db.PutItem(r.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: *ck.ID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "serverdetails",
			},
			"ServerIP": &types.AttributeValueMemberS{
				Value: *ck.IP,
			},
			"ServerName": &types.AttributeValueMemberS{
				Value: *ck.Name,
			},
			"LastUpdated": &types.AttributeValueMemberS{
				Value: time.Now().In(zone).Format(time.DateTime),
			},
		},
	})
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "server registered")
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	out, err := h.Client.db.Scan(r.Context(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var servers *[]Server
	err = attributevalue.UnmarshalListOfMaps(out.Items, &servers)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, "failed to unmarshal Dynamodb reqeust "+err.Error())
		return
	}

	if err := json.NewEncoder(w).Encode(servers); err != nil {
		WriteResponse(w, http.StatusInternalServerError, "failed to marshal response: "+err.Error())
		return
	}
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil || ck.Name == nil {
		WriteResponse(w, http.StatusBadRequest, "serverID, server name must be provided")
		return
	}

	err = ckec2.StartEC2Instance(r.Context(), h.Client.ec, ck.ID)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Server Started")
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, http.StatusBadRequest, "serverID must be provided")
		return
	}

	commands := []string{
		"sudo docker exec -i " + *ck.Name + " rcon-cli", "stop",
		"sudo aws s3 sync --delete data s3://creeperkeeper-world-data/" + *ck.Name + "/",
	}
	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{*ck.ID},
		Parameters: map[string][]string{
			"commands":         commands,
			"workingDirectory": {"/home/ec2-user"},
		},
	}
	_, err = h.Client.sc.SendCommand(r.Context(), input)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to stop minecraft server: %s", err.Error()))
		return
	}

	err = ckec2.StopEC2Instance(r.Context(), h.Client.ec, ck.ID)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Server stopping")
}

// getCommandOutput polls until the command is finished and prints its output.
func getCommandOutput(ctx context.Context, client *ssm.Client, commandID *string, instanceID *string) error {
	for {
		invocation, err := client.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(*commandID),
			InstanceId: aws.String(*instanceID),
		})
		if err != nil {
			return fmt.Errorf("error getting command invocation: %w", err)
		}

		// Check if the command is still running.
		if invocation.Status == "InProgress" || invocation.Status == "Pending" {
			fmt.Println("Command still running. Waiting 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		// Once the command has finished, print the outputs.
		fmt.Printf("Command Status: %s\n", invocation.Status)
		fmt.Println("Standard Output:")
		fmt.Println(invocation.StandardOutputContent)
		fmt.Println("Standard Error:")
		fmt.Println(invocation.StandardErrorContent)
		break
	}
	return nil
}

func WriteResponse(w http.ResponseWriter, code int, message interface{}) {
	w.WriteHeader(code)
	response := map[string]interface{}{"message": message}
	jMessage, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	w.Write(jMessage)
}
