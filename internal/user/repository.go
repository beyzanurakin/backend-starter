package user

var users = []User{
    {ID: 1, Name: "John Doe", Email: "john@example.com", Role: "user"},
    {ID: 2, Name: "Jane Doe", Email: "jane@example.com", Role: "admin"},
}

func GetAllUsers() []User {
    return users
}

func GetUserByID(id int) *User {
    for _, u := range users {
        if u.ID == id {
            return &u
        }
    }
    return nil
}
