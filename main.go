package main

import (
    "fmt"
    "log"
    "forum/db" // import the db package
    "net/http"
)

func main() {
    // Initialize the database
    err := db.Initialize()
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close() // Ensure the database is closed when the app stops

    // Your HTTP server and routes setup here

    fmt.Println("Server started on :8080")
    err = http.ListenAndServe(":8080", nil) // Replace nil with your router if needed
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}
