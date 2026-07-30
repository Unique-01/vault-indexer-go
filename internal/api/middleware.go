package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
