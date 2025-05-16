package auth

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	userModel "backend-starter/internal/user"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var u userModel.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}
	u.Password = string(hashedPassword)
	u.Role = "user" // Default role

	created := userModel.CreateUser(u)

	created.Password = "" // şifreyi gizle
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
