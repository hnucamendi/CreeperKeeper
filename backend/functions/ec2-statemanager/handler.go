package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type C struct {
	EC2 *ec2.Client
	DB  *dynamodb.Client
}

type Handler struct {
	Client *C
}

type EC2State uint8

const (
	STOP EC2State = iota
	START
	TERMINATE
	DESCRIBE
)

func NewHandler(c *C) *Handler {
	return &Handler{
		Client: c,
	}
}

type EC2RequestBody struct {
	InstanceID   string   `json:"instanceID"`
	DesiredState EC2State `json:"desiredState"`
}

func (h *Handler) manageInstanceState(w http.ResponseWriter, r *http.Request) {
	var requestBody *EC2RequestBody
	json.NewDecoder(r.Body).Decode(&requestBody)

	switch requestBody.DesiredState {
	case STOP:
		// Stop the EC2 instance
		if err := StopEC2Instance(r.Context(), h.Client.EC2, requestBody.InstanceID); err != nil {
			WriteResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteResponse(w, http.StatusOK, "Instance stopped")
		return
	case START:
		// Start the EC2 instance
		out, err := StartEC2Instance(r.Context(), h.Client.EC2, requestBody.InstanceID)
		if err != nil {
			WriteResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		WriteResponse(w, http.StatusOK, out)
		return
	case TERMINATE:
		// Terminate the EC2 instance
		if err := TerminateEC2Instance(r.Context(), h.Client.EC2, h.Client.DB, requestBody.InstanceID); err != nil {
			WriteResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		WriteResponse(w, http.StatusOK, "Instance terminated")
		return
	default:
		WriteResponse(w, http.StatusBadRequest, "Invalid desired state")
		return
	}
}
