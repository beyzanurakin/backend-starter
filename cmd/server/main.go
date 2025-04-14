package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "time"
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
    "golang.org/x/time/rate"
)

var (
    users       = map[int]User{}
    transactions = map[int]Transaction{}
    balances    = map[int]Balance{}
    rateLimiter = rate.NewLimiter(rate.Every(time.Minute), 5)
)

type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password,omitempty"`
    Role     string `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Amount    float64   `json:"amount"`
    Type      string    `json:"type"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

type Balance struct {
    UserID int     `json:"user_id"`
    Amount float64 `json:"amount"`
}

func main() {
    r := mux.NewRouter()
    r.Use(authenticationMiddleware)
    r.Use(performanceMonitoringMiddleware)
    r.Use(requestValidationMiddleware)
    r.Use(errorHandlingMiddleware)

    adminRouter := r.PathPrefix("/admin").Subrouter()
    adminRouter.Use(roleAuthorizationMiddleware("admin", adminRouter))

    r.HandleFunc("/api/v1/auth/register", registerUser).Methods("POST")
    r.HandleFunc("/api/v1/auth/login", loginUser).Methods("POST")
    r.HandleFunc("/api/v1/users", getUsers).Methods("GET")
    r.HandleFunc("/api/v1/users/{id}", getUser).Methods("GET")
    r.HandleFunc("/api/v1/users/{id}", updateUser).Methods("PUT")
    r.HandleFunc("/api/v1/users/{id}", deleteUser).Methods("DELETE")
    r.HandleFunc("/api/v1/transactions/credit", creditTransaction).Methods("POST")
    r.HandleFunc("/api/v1/transactions/debit", debitTransaction).Methods("POST")
    r.HandleFunc("/api/v1/transactions/transfer", transferTransaction).Methods("POST")
    r.HandleFunc("/api/v1/transactions/history", getTransactionHistory).Methods("GET")
    r.HandleFunc("/api/v1/transactions/{id}", getTransaction).Methods("GET")
    r.HandleFunc("/api/v1/balances/current", getCurrentBalance).Methods("GET")
    r.HandleFunc("/api/v1/balances/historical", getHistoricalBalance).Methods("GET")

    http.ListenAndServe(":8080", r)
}

func authenticationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Authorization token required", http.StatusUnauthorized)
            return
        }

        tokenString = tokenString[len("Bearer "):]
        claims := jwt.MapClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte("your_secret_key"), nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
            return
        }

        userID := claims["user_id"].(float64)
        r.Header.Set("User-ID", fmt.Sprintf("%v", userID))

        next.ServeHTTP(w, r)
    })
}

func roleAuthorizationMiddleware(requiredRole string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userRole := r.Header.Get("Role")
        if userRole != requiredRole {
            http.Error(w, "Forbidden: You do not have the necessary permissions", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func requestValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost || r.Method == http.MethodPut {
            var requestBody map[string]interface{}
            decoder := json.NewDecoder(r.Body)
            if err := decoder.Decode(&requestBody); err != nil {
                http.Error(w, "Invalid JSON format", http.StatusBadRequest)
                return
            }

            if requestBody["name"] == nil || requestBody["email"] == nil {
                http.Error(w, "Missing required fields", http.StatusBadRequest)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}

func errorHandlingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}


func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func corsMiddleware(next http.Handler) http.Handler {
    return handlers.CORS(
        handlers.AllowedOrigins([]string{"http://localhost:3000"}),
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
        handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
    )(next)
}

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !rateLimiter.Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func registerUser(w http.ResponseWriter, r *http.Request) {
    var user User
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "User %s registered successfully", user.Name)
}


func loginUser(w http.ResponseWriter, r *http.Request) {
    var user User
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    token, err := generateJWT(user)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Authorization", "Bearer "+token)
    fmt.Fprintf(w, "User %s logged in successfully", user.Name)
}


func refreshToken(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    
    fmt.Fprintf(w, "Token refreshed successfully")
}


func getUsers(w http.ResponseWriter, r *http.Request) {
    users := []User{
        {ID: 1, Name: "John Doe", Email: "john@example.com"},
        {ID: 2, Name: "Jane Doe", Email: "jane@example.com"},
    }
    
    json.NewEncoder(w).Encode(users)
}


func getUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}

    json.NewEncoder(w).Encode(user)
}


func updateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    var user User
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }


    fmt.Fprintf(w, "User %s updated successfully", id)
}


func deleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    fmt.Fprintf(w, "User %s deleted successfully", id)
}

func creditTransaction(w http.ResponseWriter, r *http.Request) {
    var transaction Transaction
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&transaction); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "Credit transaction of %f for user %d completed", transaction.Amount, transaction.UserID)
}


func debitTransaction(w http.ResponseWriter, r *http.Request) {
    var transaction Transaction
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&transaction); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "Debit transaction of %f for user %d completed", transaction.Amount, transaction.UserID)
}


func transferTransaction(w http.ResponseWriter, r *http.Request) {
    var transaction Transaction
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&transaction); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "Transfer transaction of %f from user %d completed", transaction.Amount, transaction.UserID)
}

func getTransactionHistory(w http.ResponseWriter, r *http.Request) {
    transactions := []Transaction{
        {ID: 1, UserID: 1, Amount: 100, Type: "credit"},
        {ID: 2, UserID: 1, Amount: 50, Type: "debit"},
    }
    
    json.NewEncoder(w).Encode(transactions)
}

func getTransaction(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    transaction := Transaction{ID: 1, UserID: 1, Amount: 100, Type: "credit"}

    json.NewEncoder(w).Encode(transaction)
}

func getCurrentBalance(w http.ResponseWriter, r *http.Request) {
    balance := Balance{UserID: 1, Amount: 1000.0}
    
    json.NewEncoder(w).Encode(balance)
}

func getHistoricalBalance(w http.ResponseWriter, r *http.Request) {
    historicalBalance := []Balance{
        {UserID: 1, Amount: 950.0},
        {UserID: 1, Amount: 1000.0},
    }
    
    json.NewEncoder(w).Encode(historicalBalance)
}
