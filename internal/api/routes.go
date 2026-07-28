package api

import "net/http"

func (server *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", server.handleHealth)
}
