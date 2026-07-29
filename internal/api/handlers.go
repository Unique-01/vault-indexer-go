package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unique-01/vault-indexer-go/internal/indexer"
)

func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (server *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	walletParam := query.Get("walletAddress")
	if !common.IsHexAddress(walletParam) {
		http.Error(w, "missing or invalid wallet address", http.StatusBadRequest)
		return
	}
	walletAddress := common.HexToAddress(walletParam)

	filter := EventFilter{
		WalletAddress: walletAddress,
	}

	if eventTypeParam := query.Get("eventType"); eventTypeParam != "" {
		eventType := indexer.EventType(eventTypeParam)
		if !eventType.Valid() {
			http.Error(w, "invalid eventType", http.StatusBadRequest)
			return
		}
		filter.EventType = &eventType
	}

	if cursorParam := query.Get("cursor"); cursorParam != "" {
		cursor, err := DecodeCursor(cursorParam)
		if err != nil {
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		filter.Cursor = &cursor
	}

	if limitParam := query.Get("limit"); limitParam != "" {
		limit, err := strconv.Atoi(limitParam)
		if err != nil || limit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limit
	}

	page, err := server.store.ListEvents(r.Context(), filter)
	if err != nil {
		server.logger.Error("list events failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(page); err != nil {
		server.logger.Error("encode response failed", "error", err)
	}
}
