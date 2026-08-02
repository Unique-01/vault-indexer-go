package api

import "net/http"

func (server *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", server.handleHealth)

	mux.HandleFunc("GET /events",
		server.withRequestID(
			server.logRequests(
				server.requireAuth(
					server.handleListEvents))))

	mux.HandleFunc("GET /auth/challenge",
		server.withRequestID(
			server.logRequests(
				server.handleChallenge)))

	mux.HandleFunc("POST /auth/verify",
		server.withRequestID(
			server.logRequests(
				server.handleVerify)))
}
