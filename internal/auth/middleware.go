package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-starter/internal/user"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("your_secret_key")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
			http.Error(w, "Missing or malformed token", http.StatusUnauthorized)
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("User-ID", strconv.Itoa(int(claims["user_id"].(float64))))
		r.Header.Set("User-Role", claims["role"].(string))
		next.ServeHTTP(w, r)
	})
}

func RoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("User-Role") != role {
				http.Error(w, "Forbidden: insufficient privileges", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ValidationMiddleware(requiredFields ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				var payload map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "Invalid JSON format", http.StatusBadRequest)
					return
				}
				for _, field := range requiredFields {
					if _, ok := payload[field]; !ok {
						http.Error(w, "Missing field: "+field, http.StatusBadRequest)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func PerformanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("%s %s took %v", r.Method, r.URL.Path, duration)
	})
}