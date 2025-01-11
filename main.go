package main

import (
    "fmt"
    "log"
    "forum/db"
    "net/http"
)

func main() {
    // Initialize the database
    err := db.Initialize()
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close() 

    // INSERT HTTP SERVER AND ROUTES

    fmt.Println("Server started on :8080")
    err = http.ListenAndServe(":8080", nil) 
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}
