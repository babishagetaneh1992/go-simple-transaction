package http

import (
	"account/internal/core/port"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type AccountHandler struct {
	repo           port.AccountRepository
	balanceHandler http.HandlerFunc
}

func NewAccountHandler(repo port.AccountRepository, balanceHandler http.HandlerFunc) *AccountHandler {
	return &AccountHandler{repo: repo, balanceHandler: balanceHandler}
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	acc, err := h.repo.Create(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(acc)

}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	acc, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(acc)

}

func (h *AccountHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	if h.balanceHandler != nil {
		r.Get("/{id}/balance", h.balanceHandler)
	}
	return r
}
