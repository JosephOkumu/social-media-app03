package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

// global variable for database connection
var DB *sql.DB

// Initialize initializes the database connection and applies the schema
func Initialize() error {
    var err error
    DB, err = sql.Open("sqlite3", "./forum.db")
    if err != nil {
        return fmt.Errorf("failed to open database: %v", err)
    }

    // Ensure the database is accessible
    err = DB.Ping()
    if err != nil {
        return fmt.Errorf("failed to ping database: %v", err)
    }

    // Run the schema SQL to create tables
    err = applySchema()
    if err != nil {
        return fmt.Errorf("failed to apply schema: %v", err)
    }

    log.Println("Database initialized successfully")
    return nil
}


