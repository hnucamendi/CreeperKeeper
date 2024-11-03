package main

import (
	"bytes"
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
	"github.com/hnucamendi/jwt-go/jwt"
)

type EC2State int

const (
	STOP EC2State = iota
	START
	TERMINATE
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

func (h *Handler) AddInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Landed in AddInstance Route")
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
	log.Println("Landed in GetInstances Route")
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
	log.Println("Landed in StartServer Route")
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

	token, err := getToken(h.Client.j, h.Client.Client, h.Client.sc)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	body := map[string]interface{}{
		"instanceID":   ck.InstanceID,
		"desiredState": START,
	}

	jb, err := json.Marshal(body)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	ec2URL := "https://statemanager.creeperkeeper.com"
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ec2", ec2URL), bytes.NewBuffer(jb))
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	log.Println("Starting EC2 instance")
	resp, err := h.Client.Client.Do(req)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to start ec2 instance: %s", err.Error()))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error starting server", resp.StatusCode, err)
		WriteResponse(w, http.StatusInternalServerError, "Error starting server")
		return
	}

	log.Println("Starting Minecraft server")
	commands := []string{"pwd", `tmux new -d -s minecraft "echo -e 'yes' | ./start.sh"`}

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

	WriteResponse(w, http.StatusOK, "Server starting")
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

	log.Println("Stopping Minecraft server")
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

	token, err := getToken(h.Client.j, h.Client.Client, h.Client.sc)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	body := map[string]interface{}{
		"instanceID":   ck.InstanceID,
		"desiredState": STOP,
	}

	jb, err := json.Marshal(body)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	ec2URL := "https://statemanager.creeperkeeper.com"
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/ec2", ec2URL), bytes.NewBuffer(jb))
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	log.Println("Stopping EC2 instance")
	resp, err := h.Client.Client.Do(req)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error stopping server", resp.StatusCode, err)
		WriteResponse(w, http.StatusInternalServerError, "Error stopping server")
		return
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

func getToken(j *jwt.JWT, hc *http.Client, sc *ssm.Client) (string, error) {
	clientID, clientSecret, audience, err := loadEnvVars(context.Background(), sc)
	if err != nil {
		return "", err
	}

	token, err := j.GenerateToken(hc, &jwt.CreateJWTConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Audience:     audience,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}
