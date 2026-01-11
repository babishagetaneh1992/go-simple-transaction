package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	//"strconv"

	"github.com/go-chi/chi"
)

type TransactionHandler struct {
	service TransactionService
}

func NewTransactionHandler(service TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	// Implementation of Deposit handler
	var req struct {
		AccountID int64  `json:"account_id"`
		Amount    int64  `json:"amount"`
		Note      string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.Deposit(r.Context(), req.AccountID, req.Amount, req.Note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}



func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	// Implementation of Withdraw handler
	var req struct {
		AccountID int64  `json:"account_id"`
		Amount    int64  `json:"amount"`
		Note      string `json:"note"`
	}


	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.Withdraw(r.Context(), req.AccountID, req.Amount, req.Note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

}




func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	// Implementation of Transfer handler
	var req struct {
		FromAccountID int64  `json:"from_account_id"`
		ToAccountID   int64  `json:"to_account_id"`
		Amount        int64  `json:"amount"`
		Note          string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	 }


	if err := h.service.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.Note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}



func (h *TransactionHandler) History(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
       http.Error(w, "invalid account id", http.StatusBadRequest)
	   return
	}

	entries, err := h.service.History(r.Context(), id)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)

}


func (h *TransactionHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/deposit", h.Deposit)
	r.Post("/withdraw", h.Withdraw)
	r.Post("/transfer", h.Transfer)
	//r.Get("/history/{id}",h.History )
	return r
}
