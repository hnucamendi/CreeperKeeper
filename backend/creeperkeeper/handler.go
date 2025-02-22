package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
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
	ID          *string `json:"serverID" dynamodbav:"PK"`
	SK          *string `json:"row" dynamodbav:"SK"`
	IP          *string `json:"serverIP" dynamodbav:"ServerIP"`
	Name        *string `json:"serverName" dynamodbav:"ServerName"`
	LastUpdated *string `json:"lastUpdated" dynamodbav:"LastUpdated"`
	IsRunning   *bool   `json:"isRunning" dynamodbav:"IsRunning"`
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
		WriteResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, r, http.StatusBadRequest, "serverID required for registering new server")
		return
	}

	if ck.SK == nil {
		WriteResponse(w, r, http.StatusBadRequest, "serverID required for registering new server")
		return
	}

	if ck.IP == nil {
		WriteResponse(w, r, http.StatusBadRequest, "IP required for registering new server")
	}

	if ck.Name == nil {
		WriteResponse(w, r, http.StatusBadRequest, "server name is required for registering new server")
	}

	if ck.IsRunning == nil {
		WriteResponse(w, r, http.StatusBadRequest, "server name is required for registering new server")
	}

	if ck.LastUpdated == nil {
		WriteResponse(w, r, http.StatusBadRequest, "server name is required for registering new server")
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
				Value: *ck.LastUpdated,
			},
			"IsRunning": &types.AttributeValueMemberBOOL{
				Value: *ck.IsRunning,
			},
		},
	})
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, r, http.StatusOK, "server registered")
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	out, err := h.Client.db.Scan(r.Context(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var servers *[]Server
	err = attributevalue.UnmarshalListOfMaps(out.Items, &servers)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, "failed to unmarshal Dynamodb reqeust "+err.Error())
		return
	}

  WriteResponse(w, r, http.StatusOK, servers)
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, r, http.StatusBadRequest, "serverID must be provided")
		return
	}

	err = ckec2.StartEC2Instance(r.Context(), h.Client.ec, ck.ID)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, r, http.StatusOK, "Server Started")
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, r, http.StatusBadRequest, "serverID must be provided")
		return
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: *ck.ID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "serverdetails",
			},
		},
	}
	out, err := h.Client.db.GetItem(r.Context(), input)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, "failed to set timezone"+err.Error())
		return
	}
	lastUpdated := time.Now().In(zone).Format(time.DateTime)

	var server *Server
	err = attributevalue.UnmarshalMap(out.Item, &server)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, "failed to unmarshal Dynamodb reqeust "+err.Error())
		return
	}

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
				Value: *server.IP,
			},
			"ServerName": &types.AttributeValueMemberS{
				Value: *server.Name,
			},
			"LastUpdated": &types.AttributeValueMemberS{
				Value: lastUpdated,
			},
			"IsRunning": &types.AttributeValueMemberBOOL{
				Value: false,
			},
		},
	})
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	commands := []string{
		"sudo docker exec -i " + *server.Name + " rcon-cli", "stop",
		"sudo aws s3 sync --delete data s3://creeperkeeper-world-data/" + *server.Name + "/",
	}
	cmdInput := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{*ck.ID},
		CloudWatchOutputConfig: &ssmTypes.CloudWatchOutputConfig{
			CloudWatchOutputEnabled: true,
			CloudWatchLogGroupName:  aws.String("/aws/lambda/creeperkeeper"),
		},
		Parameters: map[string][]string{
			"commands":         commands,
			"workingDirectory": {"/home/ec2-user"},
		},
	}
	_, err = h.Client.sc.SendCommand(r.Context(), cmdInput)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = ckec2.StopEC2Instance(r.Context(), h.Client.ec, ck.ID)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, r, http.StatusOK, "Server stopping")
}

func generateETag[T any](data T) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return `W/"` + hex.EncodeToString(hash[:]) + `"`
}

func WriteResponse[T any](w http.ResponseWriter, r *http.Request, code int, message T) {
	etag := generateETag(message)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)

	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	json.NewEncoder(w).Encode(message)
}
