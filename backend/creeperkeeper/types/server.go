package types

import (
	"encoding/json"
	"io"
)

type Server struct {
	ID          *string `json:"serverID" dynamodbav:"PK"`
	SK          *string `json:"row" dynamodbav:"SK"`
	IP          *string `json:"serverIP" dynamodbav:"ServerIP"`
	Name        *string `json:"serverName" dynamodbav:"ServerName"`
	LastUpdated *string `json:"lastUpdated" dynamodbav:"LastUpdated"`
	IsRunning   *bool   `json:"isRunning" dynamodbav:"IsRunning"`
}

func (ck *Server) UnmarshallRequest(b io.ReadCloser) error {
	err := json.NewDecoder(b).Decode(&ck)
	if err != nil {
		return err
	}

	return nil
}
