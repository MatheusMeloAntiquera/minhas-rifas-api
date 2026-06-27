package raffle

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /raffles", h.Create)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("falha ao decodificar corpo da requisição", "error", err)
		h.logger.Error("falha ao decodificar corpo da requisição 2", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisição inválido"})
		return
	}

	raffle, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao criar rifa", "error", err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, raffle)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
