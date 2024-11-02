package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Handler struct {
	Client *ssm.Client
}

func NewHandler(sc *ssm.Client) *Handler {
	return &Handler{
		Client: sc,
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

	_, err = h.Client.SendCommand(r.Context(), input)
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

	_, err = h.Client.SendCommand(r.Context(), input)
	if err != nil {
		WriteResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteResponse(w, http.StatusOK, "Server stopping")
}

func WriteResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write([]byte(message))
}
