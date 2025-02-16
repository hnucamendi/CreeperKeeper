package main

import (
	"net/http"
)

func loadRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /creeperkeeper/server/register", h.RegisterServer)
	mux.HandleFunc("GET /creeperkeeper/server/list", h.ListServers)
	mux.HandleFunc("POST /creeperkeeper/server/start", h.StartServer)
	mux.HandleFunc("POST /creeperkeeper/server/stop", h.StopServer)
	mux.HandleFunc("GET /creeperkeeper/server/ping", h.Ping)
}
