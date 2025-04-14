package user

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)
    msg := Register(user)
    w.Write([]byte(msg))
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)
    msg := Login(user.Email, user.Password)
    w.Write([]byte(msg))
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(GetAllUsers())
}

func GetUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, _ := strconv.Atoi(idStr)

    user := GetUserByID(id)
    if user == nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(user)
}
