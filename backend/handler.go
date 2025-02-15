package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

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

type CreeperKeeper struct {
	InstanceID string `json:"instanceID"`
}

func (ck *CreeperKeeper) unmarshallRequest(b io.ReadCloser) error {
	err := json.NewDecoder(b).Decode(&ck)
	if err != nil {
		return err
	}

	return nil
}

// Adds EC2 instance details to DynamoDB
func (h *Handler) AddInstance(w http.ResponseWriter, r *http.Request) {
	ck := &CreeperKeeper{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.InstanceID == "" {
		WriteResponse(w, http.StatusBadRequest, "instance_id must be provided")
		return
	}

	_, err = h.Client.db.PutItem(r.Context(), &dynamodb.PutItemInput{
		TableName: aws.String("CreeperKeeper"),
		Item: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: ck.InstanceID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "instance",
			},
		},
	})
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Instance added")
}

func (h *Handler) GetInstances(w http.ResponseWriter, r *http.Request) {
	out, err := h.Client.db.Scan(r.Context(), &dynamodb.ScanInput{
		TableName: aws.String("CreeperKeeper"),
	})
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	instances := []string{}
	for _, item := range out.Items {
		instances = append(instances, item["PK"].(*types.AttributeValueMemberS).Value)
	}

	WriteResponse(w, http.StatusOK, instances)
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	ck := &CreeperKeeper{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.InstanceID == "" {
		WriteResponse(w, http.StatusBadRequest, "instance_id must be provided")
		return
	}

	instances, err := ckec2.StartEC2Instance(context.Background(), h.Client.ec2, ck.InstanceID)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
	}

	b, err := json.Marshal(instances)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to marshal instance list"))
	}

	commands := []string{`tmux new -d -s minecraft "echo -e 'yes' | ./start.sh"`}

	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{ck.InstanceID},
		Parameters: map[string][]string{
			"commands":         commands,
			"workingDirectory": {"/home/ec2-user/Minecraft"},
		},
	}

	_, err = h.Client.sc.SendCommand(r.Context(), input)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to start minecraft server: %s", err.Error()))
		return
	}

	WriteResponse(w, http.StatusOK, string(b))
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	ck := &CreeperKeeper{}
	err := ck.unmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if ck.InstanceID == "" {
		WriteResponse(w, http.StatusBadRequest, "instance_id must be provided")
		return
	}

	log.Println("Instance ID:", ck.InstanceID)

	commands := []string{"tmux send-keys -t minecraft 'C-c'"}

	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{ck.InstanceID},
		Parameters: map[string][]string{
			"commands":         commands,
			"workingDirectory": {"/home/ec2-user/Minecraft"},
		},
	}

	_, err = h.Client.sc.SendCommand(r.Context(), input)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to stop minecraft server: %s", err.Error()))
		return
	}

	err = ckec2.StopEC2Instance(context.Background(), h.Client.ec2, ck.InstanceID)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to stop minecraft server %s", err.Error()))
	}

	WriteResponse(w, http.StatusOK, "Server stopping")
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
