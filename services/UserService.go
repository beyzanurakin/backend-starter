package main

import (
    "database/sql"
    "errors"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "time"
)

type UserServiceImpl struct {
    db *sql.DB
    jwtSecret string
}

type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password,omitempty"`
    Role     string `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

func NewUserService(db *sql.DB, secret string) *UserServiceImpl {
    return &UserServiceImpl{db: db, jwtSecret: secret}
}

func HashPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

func (s *UserServiceImpl) RegisterUser(name, email, password, role string) error {
    if name == "" || email == "" || password == "" || role == "" {
        return errors.New("all fields are required")
    }

    hashedPassword, err := HashPassword(password)
    if err != nil {
        return err
    }

    _, err = s.db.Exec("INSERT INTO users (name, email, password, role, created_at) VALUES (?, ?, ?, ?, ?)",
        name, email, hashedPassword, role, time.Now())
    if err != nil {
        return err
    }
    return nil
}

func (s *UserServiceImpl) AuthenticateUser(email, password string) (string, error) {
    var user User
    err := s.db.QueryRow("SELECT id, name, email, password, role FROM users WHERE email = ?", email).Scan(
        &user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
    if err != nil {
        return "", errors.New("invalid email or password")
    }

    if !CheckPassword(user.Password, password) {
        return "", errors.New("invalid email or password")
    }

    return s.GenerateJWT(user)
}

func (s *UserServiceImpl) GenerateJWT(user User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email": user.Email,
        "role": user.Role,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}

func (s *UserServiceImpl) AuthorizeRole(userID int, requiredRole string) (bool, error) {
    var role string
    err := s.db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
    if err != nil {
        return false, err
    }
    return role == requiredRole, nil
}

func main() {
    db, _ := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/mydb")
    defer db.Close()
    
    userService := NewUserService(db, "my_secret_key")
    
    err := userService.RegisterUser("John Doe", "john@example.com", "password123", "admin")
    if err != nil {
        fmt.Println("Registration failed:", err)
    }
    
    token, err := userService.AuthenticateUser("john@example.com", "password123")
    if err != nil {
        fmt.Println("Authentication failed:", err)
    } else {
        fmt.Println("JWT Token:", token)
    }
}
