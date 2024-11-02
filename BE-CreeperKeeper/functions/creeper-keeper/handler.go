package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

	response, err := json.Marshal(instances)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, string(response))
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
		WriteResponse(w, http.StatusInternalServerError, err.Error())
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

	commands := []string{"tmux send-keys -t minecraft 'C-c'"}

	input := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{ck.InstanceID},
		Parameters: map[string][]string{
			"commands": commands,
		},
	}

	_, err = h.Client.sc.SendCommand(r.Context(), input)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Server stopping")
}

func WriteResponse(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	message = fmt.Sprintf(`{"message": %q}`, message)
	jMessage, err := json.Marshal(message)
	if err != nil {
		w.Write([]byte(`{"message": "Internal Server Error"}`))
		return
	}
	w.Write([]byte(jMessage))
}
