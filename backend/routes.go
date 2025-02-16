package main

import (
	"net/http"
)

func loadRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /creeperkeeper/start", h.StartServer)
	mux.HandleFunc("POST /creeperkeeper/stop", h.StopServer)
	mux.HandleFunc("POST /creeperkeeper/add", h.AddInstance)
	mux.HandleFunc("GET /creeperkeeper/instances", h.GetInstances)
}
