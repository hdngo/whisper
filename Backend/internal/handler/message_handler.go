package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hdngo/whisper/internal/repository"
)

type MessageHandler struct {
	msgRepo *repository.MessageRepository
}

func NewMessageHandler(msgRepo *repository.MessageRepository) *MessageHandler {
	return &MessageHandler{msgRepo: msgRepo}
}

func (h *MessageHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit

	if limitStr != "" {
		parseLimit, err := strconv.Atoi(limitStr)
		if err == nil && parseLimit > 0 && parseLimit <= 100 {
			limit = parseLimit
		}
	}

	messages, err := h.msgRepo.GetRecent(r.Context(), limit)
	if err != nil {
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessageHandler) GetMessagesBefore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	beforeID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid message ID", http.StatusBadRequest)
		return
	}

	limit := 50 // Default limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	messages, err := h.msgRepo.GetMessagesBefore(r.Context(), beforeID, limit)
	if err != nil {
		http.Error(w, "failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
