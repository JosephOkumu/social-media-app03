package post

import (
	"database/sql"
	"testing"

	"forum/db"

	_ "github.com/mattn/go-sqlite3" // Required for in-memory SQLite
)

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create necessary tables
	setupSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE post_reactions (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			reaction_type TEXT NOT NULL CHECK(reaction_type IN ('LIKE', 'DISLIKE')),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			PRIMARY KEY (post_id, user_id)
		);
	`
	_, err = testDB.Exec(setupSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Replace global DB with test DB
	db.DB = testDB

	return testDB
}

// TestPostReactions ensures that adding, updating, and deleting reactions works correctly
func TestPostReactions(t *testing.T) {
	// Setup test database
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insert a test user
	_, err := testDB.Exec(`INSERT INTO users (id, username) VALUES (1, 'testuser')`)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Test: Add a new reaction
	_, err = testDB.Exec(`INSERT INTO post_reactions (post_id, user_id, reaction_type) VALUES (?, ?, ?)`, 1, 1, "LIKE")
	if err != nil {
		t.Fatalf("Failed to insert reaction: %v", err)
	}

	// Verify insertion
	var reactionType string
	err = testDB.QueryRow(`SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?`, 1, 1).Scan(&reactionType)
	if err != nil {
		t.Fatalf("Failed to fetch reaction: %v", err)
	}
	if reactionType != "LIKE" {
		t.Errorf("Expected reaction 'LIKE', got '%s'", reactionType)
	}

	// Test: Update reaction
	_, err = testDB.Exec(`UPDATE post_reactions SET reaction_type = ? WHERE post_id = ? AND user_id = ?`, "DISLIKE", 1, 1)
	if err != nil {
		t.Fatalf("Failed to update reaction: %v", err)
	}

	// Verify update
	err = testDB.QueryRow(`SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?`, 1, 1).Scan(&reactionType)
	if err != nil {
		t.Fatalf("Failed to fetch updated reaction: %v", err)
	}
	if reactionType != "DISLIKE" {
		t.Errorf("Expected reaction 'DISLIKE', got '%s'", reactionType)
	}

	// Test: Remove reaction
	_, err = testDB.Exec(`DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?`, 1, 1)
	if err != nil {
		t.Fatalf("Failed to delete reaction: %v", err)
	}

	// Verify deletion
	err = testDB.QueryRow(`SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?`, 1, 1).Scan(&reactionType)
	if err == nil {
		t.Errorf("Expected no reaction, but found '%s'", reactionType)
	} else if err != sql.ErrNoRows {
		t.Fatalf("Unexpected error during deletion verification: %v", err)
	}
}
