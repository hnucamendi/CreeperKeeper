package main

import (
	"net/http"
)

func loadRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /ck/start", h.StartServer)
	mux.HandleFunc("POST /ck/stop", h.StopServer)
	mux.HandleFunc("POST /ck/add", h.AddInstance)
	mux.HandleFunc("GET /ck/instances", h.GetInstances)
	mux.HandleFunc("GET /ck/test", h.Test)
}
