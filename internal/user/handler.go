package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type TransferRequest struct {
	FromUserID int     `json:"from_user_id"`
	ToUserID   int     `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

func CreditHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	res := Credit(req.UserID, req.Amount)
	json.NewEncoder(w).Encode(res)
}

func DebitHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	res, ok := Debit(req.UserID, req.Amount)
	if !ok {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	res, ok := Transfer(req.FromUserID, req.ToUserID, req.Amount)
	if !ok {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("User-ID")
	userID, _ := strconv.Atoi(userIDStr)
	balance := GetBalance(userID)
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

func AllTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	list := GetAllTransactions()
	json.NewEncoder(w).Encode(list)
}