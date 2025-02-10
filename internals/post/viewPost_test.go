package post

import (
	"database/sql"
	"testing"
	"time"

	"forum/db"

	_ "github.com/mattn/go-sqlite3"
)

func TestFetchPostFromDB(t *testing.T) {
	// Setup test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create required tables
	setupSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image TEXT,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE post_reactions (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			PRIMARY KEY (post_id, user_id)
		);
	`
	_, err = testDB.Exec(setupSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	testData := `
		INSERT INTO users (id, username) VALUES 
			(1, 'testuser1'),
			(2, 'testuser2');

		INSERT INTO posts (id, title, content, image, user_id, created_at) VALUES
			(1, 'Test Post 1', 'Content of test post 1', 'image1.jpg', 1, ?),
			(2, 'Test Post 2', 'Content of test post 2', NULL, 2, ?);

		INSERT INTO comments (post_id, user_id, content, created_at) VALUES
			(1, 2, 'Comment on post 1', ?),
			(1, 1, 'Another comment on post 1', ?);

		INSERT INTO post_reactions (post_id, user_id, reaction_type) VALUES
			(1, 2, 'LIKE'),
			(1, 1, 'LIKE'),
			(2, 1, 'DISLIKE');
	`
	_, err = testDB.Exec(testData, currentTime, currentTime, currentTime, currentTime)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Temporarily replace the global DB with our test DB
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	// Test cases
	tests := []struct {
		name          string
		postID        string
		userID        int64
		expectedTitle string
		expectedError bool
	}{
		{
			name:          "Fetch existing post with image",
			postID:        "1",
			userID:        1,
			expectedTitle: "Test Post 1",
			expectedError: false,
		},
		{
			name:          "Fetch existing post without image",
			postID:        "2",
			userID:        1,
			expectedTitle: "Test Post 2",
			expectedError: false,
		},
		{
			name:          "Fetch non-existent post",
			postID:        "999",
			userID:        1,
			expectedError: true,
		},
		{
			name:          "Fetch post with invalid ID",
			postID:        "invalid",
			userID:        1,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := fetchPostFromDB(tt.postID, tt.userID)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify post data
			if post.Title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, post.Title)
			}

			// Verify reaction counts
			if tt.postID == "1" {
				if post.Likes != 2 {
					t.Errorf("Expected 2 likes, got %d", post.Likes)
				}
				if post.Dislikes != 0 {
					t.Errorf("Expected 0 dislikes, got %d", post.Dislikes)
				}
			}

			// Verify comment count
			if tt.postID == "1" {
				if post.CommentCount != 2 {
					t.Errorf("Expected 2 comments, got %d", post.CommentCount)
				}
			}
		})
	}
}

func TestFetchPosts(t *testing.T) {
	// Setup test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tables including comments table
	setupSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image TEXT,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		CREATE TABLE post_reactions (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			PRIMARY KEY (post_id, user_id)
		);
	`
	_, err = testDB.Exec(setupSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data including comments
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err = testDB.Exec(`
		INSERT INTO users (id, username) VALUES 
			(1, 'testuser1'), 
			(2, 'testuser2');

		INSERT INTO posts (id, title, content, user_id, created_at) VALUES
			(1, 'Test Post 1', 'Content 1', 1, ?),
			(2, 'Test Post 2', 'Content 2', 2, ?);

		INSERT INTO comments (post_id, user_id, content, created_at) VALUES
			(1, 2, 'Comment on post 1', ?),
			(1, 1, 'Another comment on post 1', ?),
			(2, 1, 'Comment on post 2', ?);

		INSERT INTO post_reactions (post_id, user_id, reaction_type) VALUES
			(1, 2, 'LIKE'),
			(2, 1, 'DISLIKE');
	`, currentTime, currentTime, currentTime, currentTime, currentTime)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Replace global DB
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	tests := []struct {
		name          string
		userID        int64
		expectedCount int
		checkFunc     func(*testing.T, []Post) bool
	}{
		{
			name:          "Fetch all posts",
			userID:        1,
			expectedCount: 2,
			checkFunc: func(t *testing.T, posts []Post) bool {
				if len(posts) != 2 {
					return false
				}
				// Check specific post attributes
				for _, post := range posts {
					if post.Title == "" || post.Content == "" {
						t.Error("Post missing required fields")
						return false
					}
				}
				// Check comment counts
				foundPost1 := false
				foundPost2 := false
				for _, post := range posts {
					if post.ID == 1 {
						foundPost1 = true
						if post.CommentCount != 2 {
							t.Errorf("Expected 2 comments for post 1, got %d", post.CommentCount)
							return false
						}
					}
					if post.ID == 2 {
						foundPost2 = true
						if post.CommentCount != 1 {
							t.Errorf("Expected 1 comment for post 2, got %d", post.CommentCount)
							return false
						}
					}
				}
				if !foundPost1 || !foundPost2 {
					t.Error("Not all expected posts were found")
					return false
				}
				return true
			},
		},
		{
			name:          "Check reaction counts",
			userID:        1,
			expectedCount: 2,
			checkFunc: func(t *testing.T, posts []Post) bool {
				for _, post := range posts {
					if post.ID == 1 && post.Likes != 1 {
						t.Errorf("Expected 1 like for post 1, got %d", post.Likes)
						return false
					}
					if post.ID == 2 && post.Dislikes != 1 {
						t.Errorf("Expected 1 dislike for post 2, got %d", post.Dislikes)
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := FetchPosts(tt.userID)
			if err != nil {
				t.Fatalf("FetchPosts returned unexpected error: %v", err)
			}

			if len(posts) != tt.expectedCount {
				t.Errorf("Expected %d posts, got %d", tt.expectedCount, len(posts))
			}

			if tt.checkFunc != nil {
				if !tt.checkFunc(t, posts) {
					t.Error("Check function failed")
				}
			}
		})
	}
}
