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
)

func main() {
    db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Run migrations
    err = migrate(db)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    fmt.Println("Migrations completed successfully")
}

// Migration function
func migrate(db *sql.DB) error {
    migrationsDir := "../migrations"
    files, err := os.ReadDir(migrationsDir)
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %v", err)
    }

    // Sort migrations by name
    sortedFiles := sortMigrationFiles(files)

    // Apply migrations
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

// Run a single migration file
func runMigration(db *sql.DB, filePath string) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read migration file %s: %v", filePath, err)
    }

    // Run the SQL content
    _, err = db.Exec(string(content))
    if err != nil {
        return fmt.Errorf("failed to execute migration file %s: %v", filePath, err)
    }

    return nil
}

// Sort migration files by name (assuming the name format is "YYYYMMDD_HHMMSS_migration.sql")
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
