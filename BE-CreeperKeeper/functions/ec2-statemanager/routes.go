package main

import (
	"net/http"
)

func loadRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /ec2-statemanager-stage/ec2", h.manageInstanceState)
}
