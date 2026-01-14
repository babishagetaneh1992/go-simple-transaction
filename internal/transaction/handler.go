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
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON payload")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_AMOUNT", "Amount must be greater than zero")
		return
	}


	if err := h.service.Deposit(r.Context(), req.AccountID, req.Amount, req.Note); err != nil {
		respondError(w, http.StatusBadRequest, "DEPOSIT_FAILED", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, SuccessResponse{
		Status:  "success",
		Message: "Deposit completed successfully",
	})
}



func (h *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID int64  `json:"account_id"`
		Amount    int64  `json:"amount"`
		Note      string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON payload")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_AMOUNT", "Withdrawal amount must be greater than zero")
		return
	}

	if err := h.service.Withdraw(r.Context(), req.AccountID, req.Amount, req.Note); err != nil {
		code := "WITHDRAW_FAILED"
		if err.Error() == "insufficient funds" {
			code = "INSUFFICIENT_FUNDS"
		}
		respondError(w, http.StatusBadRequest, code, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, SuccessResponse{
		Status:  "success",
		Message: "Withdrawal completed successfully",
	})
}




func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromAccountID int64  `json:"from_account_id"`
		ToAccountID   int64  `json:"to_account_id"`
		Amount        int64  `json:"amount"`
		Note          string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON payload")
		return
	}

	if req.FromAccountID == req.ToAccountID {
		respondError(w, http.StatusBadRequest, "INVALID_TRANSFER", "Cannot transfer to the same account")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "INVALID_AMOUNT", "Transfer amount must be greater than zero")
		return
	}

	if err := h.service.Transfer(
		r.Context(),
		req.FromAccountID,
		req.ToAccountID,
		req.Amount,
		req.Note,
	); err != nil {
		respondError(w, http.StatusBadRequest, "TRANSFER_FAILED", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, SuccessResponse{
		Status:  "success",
		Message: "Transfer completed successfully",
	})
}




func (h *TransactionHandler) History(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID must be a number")
		return
	}

	entries, err := h.service.History(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		return
	}

	respondJSON(w, http.StatusOK, SuccessResponse{
		Status:  "success",
		Message: "Transaction history retrieved successfully",
		Data:    entries,
	})
}






func (h *TransactionHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/deposit", h.Deposit)
	r.Post("/withdraw", h.Withdraw)
	r.Post("/transfer", h.Transfer)
	r.Get("/history/{id}",h.History )
    r.Get("/{id}/balance", h.Balance)
	return r
}


func (h *TransactionHandler) Balance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "ivalid id", http.StatusBadRequest)
		return
	}

	balance, err := h.service.Balance(r.Context(), id)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{
		"balance": balance,
	})
}









type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{Status: "error"}
	resp.Error.Code = code
	resp.Error.Message = message
	respondJSON(w, status, resp)
}
