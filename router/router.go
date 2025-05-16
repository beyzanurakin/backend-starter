package router

import (
	"net/http"

	"github.com/gorilla/mux"

	authHandler "backend-starter/internal/auth"
	userHandler "backend-starter/internal/user"
	trxHandler "backend-starter/internal/transaction"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/auth/register", authHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", authHandler.LoginUser).Methods("POST")
	r.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(authHandler.AuthMiddleware)
	api.Use(authHandler.ErrorMiddleware)
	api.Use(authHandler.PerformanceMiddleware)

	api.HandleFunc("/users", userHandler.GetAllUsersHandler).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.GetUserHandler).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUserHandler).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUserHandler).Methods("DELETE")

	api.HandleFunc("/transactions/credit", trxHandler.CreditHandler).Methods("POST")
	api.HandleFunc("/transactions/debit", trxHandler.DebitHandler).Methods("POST")
	api.HandleFunc("/transactions/transfer", trxHandler.TransferHandler).Methods("POST")
	api.HandleFunc("/transactions/balance", trxHandler.BalanceHandler).Methods("GET")
	api.HandleFunc("/transactions/history", trxHandler.AllTransactionsHandler).Methods("GET")

	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(authHandler.RoleMiddleware("admin"))
	admin.HandleFunc("/users", userHandler.GetAllUsersHandler).Methods("GET")

	return r
}
