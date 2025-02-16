package main

import (
	"net/http"
)

func loadRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /creeperkeeper/instance/register", h.RegisterServer)
	mux.HandleFunc("GET /creeperkeeper/instance/list", h.ListServers)
	mux.HandleFunc("POST /creeperkeeper/instance/start", h.StartServer)
	mux.HandleFunc("POST /creeperkeeper/instance/stop", h.StopServer)
	mux.HandleFunc("GET /creeperkeeper/instance/ping", h.Ping)
}
