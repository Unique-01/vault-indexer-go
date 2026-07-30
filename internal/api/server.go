package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/unique-01/vault-indexer-go/internal/auth"
)

type Server struct {
	logger *slog.Logger
	addr   string
	store  EventReader
	auth   *auth.Service
}

func New(logger *slog.Logger, addr string, store EventReader,auth *auth.Service) *Server {
	return &Server{
		logger: logger,
		addr:   addr,
		store:  store,
		auth: auth,
	}
}

func (server *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	server.registerRoutes(mux)

	httpServer := &http.Server{
		Addr:    server.addr,
		Handler: mux,
	}

	errChan := make(chan error, 1)

	go func() {
		server.logger.Info("Api Server listening", "Addr", server.addr)

		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		server.logger.Info("Api server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
