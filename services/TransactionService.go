package services

import (
    "database/sql"
    "errors"
    "fmt"
    "time"
)

type TransactionService struct {
    db *sql.DB
}

type Transaction struct {
    ID        int       `json:"id"`
    FromUser  int       `json:"from_user"`
    ToUser    int       `json:"to_user,omitempty"`
    Amount    float64   `json:"amount"`
    Type      string    `json:"type"` 
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

func NewTransactionService(db *sql.DB) *TransactionService {
    return &TransactionService{db: db}
}

func (s *TransactionService) Credit(userID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("UPDATE balances SET amount = amount + ? WHERE user_id = ?", amount, userID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("INSERT INTO transactions (user_id, amount, type, status, created_at) VALUES (?, ?, ?, ?, ?)",
        userID, amount, "credit", "completed", time.Now())
    if err != nil {
        return err
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

func (s *TransactionService) Debit(userID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var balance float64
    err = tx.QueryRow("SELECT amount FROM balances WHERE user_id = ?", userID).Scan(&balance)
    if err != nil {
        return err
    }

    if balance < amount {
        return errors.New("insufficient balance")
    }

    _, err = tx.Exec("UPDATE balances SET amount = amount - ? WHERE user_id = ?", amount, userID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("INSERT INTO transactions (user_id, amount, type, status, created_at) VALUES (?, ?, ?, ?, ?)",
        userID, amount, "debit", "completed", time.Now())
    if err != nil {
        return err
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

func (s *TransactionService) Transfer(fromUserID, toUserID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var balance float64
    err = tx.QueryRow("SELECT amount FROM balances WHERE user_id = ?", fromUserID).Scan(&balance)
    if err != nil {
        return err
    }

    if balance < amount {
        return errors.New("insufficient balance")
    }

    _, err = tx.Exec("UPDATE balances SET amount = amount - ? WHERE user_id = ?", amount, fromUserID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("UPDATE balances SET amount = amount + ? WHERE user_id = ?", amount, toUserID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("INSERT INTO transactions (user_id, amount, type, status, created_at) VALUES (?, ?, ?, ?, ?)",
        fromUserID, amount, "debit", "completed", time.Now())
    if err != nil {
        return err
    }

    _, err = tx.Exec("INSERT INTO transactions (user_id, amount, type, status, created_at) VALUES (?, ?, ?, ?, ?)",
        toUserID, amount, "credit", "completed", time.Now())
    if err != nil {
        return err
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}
