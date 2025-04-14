package user

func Register(user User) string {
    users = append(users, user)
    return "User registered successfully"
}

func Login(email, password string) string {
    // dummy check
    for _, u := range users {
        if u.Email == email {
            return "Login successful"
        }
    }
    return "Invalid credentials"
}
