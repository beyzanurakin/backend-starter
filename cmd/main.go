package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sort"
    _ "github.com/go-sql-driver/mysql"
    "errors"
    "time"
    "sync"
    "encoding/json"
    "sync/atomic"
)

var (
    successfulTransactions int32
    failedTransactions     int32
)

func main() {
    db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    err = migrate(db)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    fmt.Println("Migrations completed successfully")

    transactionQueue := make(chan *Transaction, 100)
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go worker(i, transactionQueue, &wg)
    }

    for i := 0; i < 10; i++ {
        transactionQueue <- &Transaction{
            ID:        i,
            UserID:    i % 3,
            Amount:    float64(i * 100),
            Type:      "credit",
            Status:    "pending",
            CreatedAt: time.Now(),
        }
    }

    close(transactionQueue)

    wg.Wait()

    fmt.Printf("Successful Transactions: %d\n", atomic.LoadInt32(&successfulTransactions))
    fmt.Printf("Failed Transactions: %d\n", atomic.LoadInt32(&failedTransactions))

}

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("name is required")
    }
    if u.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

type Transaction struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Amount    float64   `json:"amount"`
    Type      string    `json:"type"`   
    Status    string    `json:"status"` 
    CreatedAt time.Time `json:"created_at"`
}

func (t *Transaction) SetState(status string) {
    t.Status = status
}

type Balance struct {
    UserID int     `json:"user_id"`
    Amount float64 `json:"amount"`
    mu     sync.Mutex
}

func (b *Balance) Add(amount float64) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.Amount += amount
}

func (b *Balance) Subtract(amount float64) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    if b.Amount < amount {
        return errors.New("insufficient balance")
    }
    b.Amount -= amount
    return nil
}

func (b *Balance) GetBalance() float64 {
    b.mu.Lock()
    defer b.mu.Unlock()
    return b.Amount
}

func worker(id int, transactionQueue <-chan *Transaction, wg *sync.WaitGroup) {
    defer wg.Done()
    for transaction := range transactionQueue {
        fmt.Printf("Worker %d processing transaction %d\n", id, transaction.ID)
        if err := processTransaction(transaction); err != nil {
            atomic.AddInt32(&failedTransactions, 1)
            fmt.Printf("Worker %d failed to process transaction %d: %v\n", id, transaction.ID, err)
        } else {
            atomic.AddInt32(&successfulTransactions, 1)
            fmt.Printf("Worker %d successfully processed transaction %d\n", id, transaction.ID)
        }
    }
}

func processTransaction(transaction *Transaction) error {
    time.Sleep(500 * time.Millisecond)

    if transaction.ID%2 == 0 {
        return nil 
    }

    return errors.New("transaction processing failed")
}

type UserRepository interface {
    Save(user *User) error
    FindByID(id int) (*User, error)
}

type TransactionRepository interface {
    Save(transaction *Transaction) error
    FindByUserID(userID int) ([]*Transaction, error)
}

type BalanceRepository interface {
    Save(balance *Balance) error
    FindByUserID(userID int) (*Balance, error)
}

type UserService interface {
    CreateUser(user *User) error
    GetUser(id int) (*User, error)
}

type TransactionService interface {
    CreateTransaction(transaction *Transaction) error
    GetUserTransactions(userID int) ([]*Transaction, error)
}

type BalanceService interface {
    AddFunds(userID int, amount float64) error
    DeductFunds(userID int, amount float64) error
    GetBalance(userID int) (float64, error)
}

func (u *User) ToJSON() ([]byte, error) {
    return json.Marshal(u)
}

func (u *User) FromJSON(data []byte) error {
    return json.Unmarshal(data, u)
}

func (t *Transaction) ToJSON() ([]byte, error) {
    return json.Marshal(t)
}

func (t *Transaction) FromJSON(data []byte) error {
    return json.Unmarshal(data, t)
}

func (b *Balance) ToJSON() ([]byte, error) {
    b.mu.Lock()
    defer b.mu.Unlock()
    return json.Marshal(b)
}

func (b *Balance) FromJSON(data []byte) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    return json.Unmarshal(data, b)
}

func migrate(db *sql.DB) error {
    migrationsDir := "../migrations"
    files, err := os.ReadDir(migrationsDir)
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %v", err)
    }

    sortedFiles := sortMigrationFiles(files)

    for _, file := range sortedFiles {
        filePath := filepath.Join(migrationsDir, file)
        err := runMigration(db, filePath)
        if err != nil {
            return fmt.Errorf("failed to run migration '%s': %v", file, err)
        }
        fmt.Printf("Successfully applied migration: %s\n", file)
    }
    return nil
}

func runMigration(db *sql.DB, filePath string) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read migration file %s: %v", filePath, err)
    }

    _, err = db.Exec(string(content))
    if err != nil {
        return fmt.Errorf("failed to execute migration file %s: %v", filePath, err)
    }

    return nil
}

func sortMigrationFiles(files []os.DirEntry) []string {
    var fileNames []string
    for _, file := range files {
        if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
            fileNames = append(fileNames, file.Name())
        }
    }
    sort.Strings(fileNames)
    return fileNames
}
