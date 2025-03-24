package services

import (
    "database/sql"
    "errors"
    "fmt"
    "sync"
    "time"
)

type BalanceService struct {
    db *sql.DB
    mu sync.Mutex
}

type Balance struct {
    UserID int     `json:"user_id"`
    Amount float64 `json:"amount"`
}

type BalanceHistory struct {
    UserID    int       `json:"user_id"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}

func NewBalanceService(db *sql.DB) *BalanceService {
    return &BalanceService{db: db}
}

func (s *BalanceService) GetBalance(userID int) (float64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    var balance float64
    err := s.db.QueryRow("SELECT SUM(amount) FROM balance_history WHERE user_id = ?", userID).Scan(&balance)
    if err != nil {
        return 0, fmt.Errorf("could not fetch balance: %v", err)
    }

    return balance, nil
}

func (s *BalanceService) AddFunds(userID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.db.Exec("INSERT INTO balance_history (user_id, amount, created_at) VALUES (?, ?, ?)",
        userID, amount, time.Now())
    if err != nil {
        return fmt.Errorf("could not add funds: %v", err)
    }

    return nil
}

func (s *BalanceService) SubtractFunds(userID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    currentBalance, err := s.GetBalance(userID)
    if err != nil {
        return err
    }

    if currentBalance < amount {
        return errors.New("insufficient balance")
    }

    _, err = s.db.Exec("INSERT INTO balance_history (user_id, amount, created_at) VALUES (?, ?, ?)",
        userID, -amount, time.Now())
    if err != nil {
        return fmt.Errorf("could not subtract funds: %v", err)
    }

    return nil
}

func (s *BalanceService) TransferFunds(fromUserID, toUserID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be greater than zero")
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    fromUserBalance, err := s.GetBalance(fromUserID)
    if err != nil {
        return err
    }

    if fromUserBalance < amount {
        return errors.New("insufficient balance for the sender")
    }

    _, err = s.db.Exec("INSERT INTO balance_history (user_id, amount, created_at) VALUES (?, ?, ?)",
        fromUserID, -amount, time.Now())
    if err != nil {
        return fmt.Errorf("could not subtract funds from sender: %v", err)
    }

    _, err = s.db.Exec("INSERT INTO balance_history (user_id, amount, created_at) VALUES (?, ?, ?)",
        toUserID, amount, time.Now())
    if err != nil {
        return fmt.Errorf("could not add funds to receiver: %v", err)
    }

    return nil
}

func (s *BalanceService) RollbackTransaction(userID int, amount float64) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, err := s.db.Exec("INSERT INTO balance_history (user_id, amount, created_at) VALUES (?, ?, ?)",
        userID, -amount, time.Now())
    if err != nil {
        return fmt.Errorf("could not rollback transaction: %v", err)
    }

    return nil
}

func (s *BalanceService) GetOptimizedBalance(userID int) (float64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    var balance float64
    err := s.db.QueryRow("SELECT SUM(amount) FROM balance_history WHERE user_id = ?", userID).Scan(&balance)
    if err != nil {
        return 0, fmt.Errorf("could not fetch optimized balance: %v", err)
    }

    return balance, nil
}
