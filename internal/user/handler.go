package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.Create)
	mux.HandleFunc("PUT /users/{id}", h.Update)
	mux.HandleFunc("DELETE /users/{id}", h.Delete)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("falha ao decodificar corpo da requisição", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisição inválido"})
		return
	}

	user, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao criar usuário", "error", err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.logger.Error("id inválido na requisição", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
		return
	}

	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("falha ao decodificar corpo da requisição", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisição inválido"})
		return
	}

	user, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao atualizar usuário", "error", err, "id", id)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("falha ao deletar usuário", "error", err, "id", id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
