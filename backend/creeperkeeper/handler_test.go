package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hnucamendi/creeper-keeper/types"
)

type MockServerClient struct{}

func String(str string) *string { return &str }

func (m *MockServerClient) GetServerStatus(ctx context.Context, serverID string) (*string, error) {
	if serverID == "" {
		return nil, errors.New("serverID is required")
	}

	return String("running"), nil
}

func (m *MockServerClient) RegisterServer(ctx context.Context, tableName string, serverID string, serverType string, serverIP string, serverName string, serverIsRunning bool, serverLastUpdated string) (bool, error) {
	return true, nil
}

type MockHandler struct {
	Client *MockServerClient
}

func (h *MockHandler) Ping(w http.ResponseWriter, r *http.Request) {
	serverID := strings.TrimPrefix(r.URL.Path, "/ping/")
	if serverID == "" {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("missing serverID"))
		return
	}
	status, err := h.Client.GetServerStatus(r.Context(), serverID)
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("failed to get server status: "+err.Error()))
		return
	}
	writeResponse(w, r, http.StatusOK, status)
}

func (h *MockHandler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	ck := &types.Server{}
	err := ck.UnmarshallRequest(r.Body)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if ck.ID == nil || ck.IP == nil || ck.Name == nil || ck.IsRunning == nil || ck.LastUpdated == nil {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("missing serverID"))
		return
	}

	status, err := h.Client.RegisterServer(r.Context(), tableName, ck.ID, *ck.SK, )
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, errors.New("failed to register server status: "+err.Error()))
		return
	}
	writeResponse(w, r, http.StatusOK, status)
}

func BenchmarkPing(b *testing.B) {
	handler := &MockHandler{Client: &MockServerClient{}}
	server := httptest.NewServer(http.HandlerFunc(handler.Ping))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/ping/test-server", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.Ping(w, req)
	}
}

func BenchmarkRegisterServer(b *testing.B) {
	handler := &MockHandler{Client: &MockServerClient{}}
	server := httptest.NewServer(http.HandlerFunc(handler.RegisterServer))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL+"/register/test-server", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.RegisterServer(w, req)
	}
}
