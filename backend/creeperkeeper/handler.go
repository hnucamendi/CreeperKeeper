package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	// TODO: Use Dynamodb marshaller to simplify this logic
	var serverList = make([]map[string]string, len(out.Items))
	for i := range out.Items {
		serverList[i] = map[string]string{
			"serverID":   out.Items[i]["PK"].(*types.AttributeValueMemberS).Value,
			"serverIP":   out.Items[i]["ServerIP"].(*types.AttributeValueMemberS).Value,
			"serverName": out.Items[i]["ServerName"].(*types.AttributeValueMemberS).Value,
		}
	}

	WriteResponse(w, http.StatusOK, serverList)
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

	newServerIP, err := ckec2.Retry(r.Context(), func() (*string, error) {
		newServerIP, err := ckec2.StartEC2Instance(r.Context(), h.Client.ec, ck.ID)
		return newServerIP, err
	}, 10)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	commands := []string{"sudo docker start " + *ck.Name, "echo " + *ck.Name + " " + *ck.ID + " " + *ck.IP + " >> test.txt"}
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
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to start minecraft server: %s", err.Error()))
		return
	}

	WriteResponse(w, http.StatusOK, *newServerIP)
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	ck := &Server{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, http.StatusBadRequest, "instance_id must be provided")
		return
	}

	commands := []string{"sudo docker exec -i " + *ck.Name + " rcon-cli", "stop"}
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

	_, err = ckec2.Retry(r.Context(), func() (*string, error) {
		err := ckec2.StopEC2Instance(r.Context(), h.Client.ec, ck.ID)
		return nil, err
	}, 10)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Server stopping")
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {

	WriteResponse(w, http.StatusOK, "Not implemented yet")
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

func loadEnvVars(ctx context.Context, sc *ssm.Client) (clientID string, clientSecret string, audience string, err error) {
	envs := []string{"/statemanager/jwt/client_id", "/statemanager/jwt/client_secret", "/statemanager/jwt/audience"}
	out, err := sc.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          envs,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get parameters: %w", err)
	}

	for _, p := range out.Parameters {
		switch *p.Name {
		case "/statemanager/jwt/client_id":
			clientID = *p.Value
		case "/statemanager/jwt/client_secret":
			clientSecret = *p.Value
		case "/statemanager/jwt/audience":
			audience = *p.Value
		}
	}

	return clientID, clientSecret, audience, nil
}
