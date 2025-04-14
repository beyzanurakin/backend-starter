package router

import (
    "net/http"

    "github.com/gorilla/mux"
    "backend-start/internal/user"
)

func SetupRouter() *mux.Router {
    r := mux.NewRouter()

    r.HandleFunc("/api/v1/auth/register", user.RegisterUser).Methods("POST")
    r.HandleFunc("/api/v1/auth/login", user.LoginUser).Methods("POST")
    r.HandleFunc("/api/v1/users", user.GetUsers).Methods("GET")
    r.HandleFunc("/api/v1/users/{id}", user.GetUser).Methods("GET")

    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    return r
}
