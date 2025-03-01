package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hnucamendi/creeper-keeper/types"
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
	ck := &types.Server{}
	err := ck.UnmarshallRequest(r.Body)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		writeResponse(w, r, http.StatusBadRequest, errors.New("server id required for registering new server").Error())
		return
	}

	if ck.IP == nil {
		writeResponse(w, r, http.StatusBadRequest, errors.New("server ip required for registering new server").Error())
		return
	}

	if ck.Name == nil {
		writeResponse(w, r, http.StatusBadRequest, errors.New("server name is required for registering new server").Error())
		return
	}

	if ck.IsRunning == nil {
		writeResponse(w, r, http.StatusBadRequest, errors.New("server running status is required for registering new server").Error())
		return
	}

	if ck.LastUpdated == nil {
		writeResponse(w, r, http.StatusBadRequest, errors.New("server last updated date is required for registering new server").Error())
		return
	}

	h.Client.db.Client.RegisterServer(r.Context(), utils.ToString(h.Client.db.Table), utils.ToString(ck.ID), utils.ToString(ck.SK), utils.ToString(ck.IP), utils.ToString(ck.Name), utils.ToBool(ck.IsRunning), utils.ToString(ck.LastUpdated))

	writeResponse(w, r, http.StatusOK, "server registered")
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("serverID")
	if serverID == "" {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("missing serverID: "+serverID))
	}
	status, err := h.Client.compute.Client.GetServerStatus(r.Context(), serverID)
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("failed to get sesrver status: "+err.Error()))
	}

	writeResponse(w, r, http.StatusOK, status)
}

func (h *Handler) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.Client.db.Client.ListServers(r.Context(), utils.ToString(h.Client.db.Table))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	writeResponse(w, r, http.StatusOK, servers)
}

func (h *Handler) StartServer(w http.ResponseWriter, r *http.Request) {
	ck := &types.Server{}
	err := ck.UnmarshallRequest(r.Body)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		writeResponse(w, r, http.StatusBadRequest, "serverID must be provided")
		return
	}

	err = h.Client.compute.Client.StartServer(r.Context(), utils.ToString(ck.ID))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, r, http.StatusOK, "Server Started")
}

func (h *Handler) StopServer(w http.ResponseWriter, r *http.Request) {
	ck := &types.Server{}
	err := ck.UnmarshallRequest(r.Body)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil {
		writeResponse(w, r, http.StatusBadRequest, "serverID must be provided")
		return
	}

	server, err := h.Client.db.Client.ListServer(r.Context(), utils.ToString(h.Client.db.Table), utils.ToString(ck.ID))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	err = h.Client.db.Client.UpsertServer(r.Context(), utils.ToString(h.Client.db.Table), utils.ToString(ck.ID), utils.ToString(ck.IP), utils.ToString(ck.Name))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	err = h.Client.systemsmanagerClient.Client.Send(r.Context(), utils.ToString(server.ID), utils.ToString(server.Name))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
	}

	err = h.Client.compute.Client.StopServer(r.Context(), utils.ToString(ck.ID))
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, r, http.StatusOK, "Server stopping")
}

func generateETag[T any](data T) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return `W/"` + hex.EncodeToString(hash[:]) + `"`
}

func writeResponse[T any](w http.ResponseWriter, r *http.Request, code int, message T) {
	etag := generateETag(message)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)
	w.WriteHeader(code)

	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	json.NewEncoder(w).Encode(message)
}
