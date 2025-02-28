package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hnucamendi/creeper-keeper/ckec2"
	cktypes "github.com/hnucamendi/creeper-keeper/types"
	"github.com/hnucamendi/creeper-keeper/utils"
)

type Handler struct {
	Client *C
}

func NewHandler(c *C) *Handler {
	return &Handler{
		Client: c,
	}
}

// Adds EC2 instance details to DynamoDB to be used by EC2 Directly
// TODO: take measures to ensure this cannot be invoked from FE
func (h *Handler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	ck := &cktypes.Server{}
	err := ck.UnmarshallRequest(r.Body)
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

	h.Client.db.Client.RegisterServer(r.Context(), tableName, utils.ToString(ck.ID), utils.ToString(ck.SK), utils.ToString(ck.IP), utils.ToString(ck.Name), utils.ToBool(ck.IsRunning), utils.ToString(ck.LastUpdated))

	WriteResponse(w, r, http.StatusOK, "server registered")
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("serverID")
	if serverID == "" {
		WriteResponse(w, r, http.StatusInternalServerError, errors.New("missing serverID: "+serverID))
	}
	status, err := ckec2.GetServerStatus(r.Context(), h.Client.ec, &serverID)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, errors.New("failed to get sesrver status: "+err.Error()))
	}

	WriteResponse(w, r, http.StatusOK, status)
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.Client.db.Client.ListServers(r.Context(), tableName)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, servers)
	}

	WriteResponse(w, r, http.StatusOK, servers)
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	ck := &cktypes.Server{}
	err := ck.UnmarshallRequest(r.Body)
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
	ck := &cktypes.Server{}
	err := ck.UnmarshallRequest(r.Body)
	if err != nil {
		WriteResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		WriteResponse(w, r, http.StatusBadRequest, "serverID must be provided")
		return
	}

	server, err := h.Client.db.Client.ListServer(r.Context(), tableName, utils.ToString(ck.ID))
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	ok, err := h.Client.db.Client.UpsertServer(r.Context(), h.Client.db.Table, ck.ID, ck.IP, ck.Name)
	if err != nil {
		WriteResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	if !ok {
		WriteResponse(w, r, http.StatusInternalServerError, errors.Join(errors.New("upsert server failed")))
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
