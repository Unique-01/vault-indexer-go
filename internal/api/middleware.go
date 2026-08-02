package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/unique-01/vault-indexer-go/internal/auth"
)

type contextKey string

const walletContextKey contextKey = "wallet"

func (server *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		wallet, err := server.auth.VerifySession(token)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidSession) {
				http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			}
			server.logger.Error("verify session failed", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), walletContextKey, wallet)
		next(w, r.WithContext(ctx))
	}
}

func walletFromContext(ctx context.Context) (common.Address, bool) {
	wallet, ok := ctx.Value(walletContextKey).(common.Address)
	return wallet, ok
}

type loggerKey struct{}

func (server *Server) withRequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		w.Header().Set("X-Request-ID", requestID)

		requestLogger := server.logger.With("requestId", requestID)
		ctx := context.WithValue(r.Context(), loggerKey{}, requestLogger)

		next(w, r.WithContext(ctx))
	}
}

func loggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (server *Server) logRequests(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromContext(r.Context())
		start := time.Now()

		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		logger.Info("request started", "method", r.Method, "path", r.URL.Path)

		next(recorder, r)

		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"duration", time.Since(start).String(),
		)
	}
}
