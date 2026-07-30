package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/auth"
)

func (server *Server) handleChallenge(w http.ResponseWriter, r *http.Request) {
	walletParam := r.URL.Query().Get("walletAddress")
	if !common.IsHexAddress(walletParam) {
		http.Error(w, "missing or invalid wallet address", http.StatusBadRequest)
		return
	}
	wallet := common.HexToAddress(walletParam)

	message, err := server.auth.IssueChallenge(r.Context(), wallet)
	if err != nil {
		server.logger.Error("issue challenge failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

type verifyRequest struct {
	WalletAddress string `json:"walletAddress"`
	Signature     string `json:"signature"`
}

func (server *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !common.IsHexAddress(req.WalletAddress) {
		http.Error(w, "invalid or missing wallet address", http.StatusBadRequest)
		return
	}
	
	wallet := common.HexToAddress(req.WalletAddress)
	if _, err := server.auth.VerifySignature(r.Context(), wallet, req.Signature); err != nil {
		if errors.Is(err, auth.ErrInvalidNonce) || errors.Is(err, auth.ErrInvalidSignature) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		server.logger.Error("signature verify failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := server.auth.IssueSession(wallet)
	if err != nil {
		server.logger.Error("issue session failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}
