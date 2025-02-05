package auth

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"forum/db"
	"forum/internals/fails"

	"github.com/gofrs/uuid"
)

func Login(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseGlob("templates/*.html"))
    
    // Check if user is already logged in
    session := CheckIfLoggedIn(w, r)
    if session != nil {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    if r.Method == http.MethodPost {
        identifier := r.FormValue("identifier") // can either be username or email
        password := r.FormValue("password")
        ipAddress := r.RemoteAddr
        users := ReadfromDb()
        
        if users == nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read users"})
            return
        }

        // First check if the identifier exists
        var foundUser *User
        isEmail := strings.Contains(identifier, "@")
        
        for _, user := range users {
            if isEmail {
                if user.Email == identifier {
                    foundUser = &user
                    break
                }
            } else {
                if user.UserName == identifier {
                    foundUser = &user
                    break
                }
            }
        }

        // If user not found, pop specific error
        if foundUser == nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            if isEmail {
                json.NewEncoder(w).Encode(map[string]string{"error": "Email not found"})
            } else {
                json.NewEncoder(w).Encode(map[string]string{"error": "Username not found"})
            }
            return
        }

        // Check password if user is found
        if !decryptPassword(foundUser.Password, password) {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{"error": "Incorrect password"})
            return
        }

        // At this point, both identifier and password are correct
        if oldsession, ok := store.GetSessionByUserId(foundUser.ID); ok {
            store.DeleteSession(oldsession.ID)
        }

        session := store.CreateSession(foundUser.ID, foundUser.UserName, ipAddress)
        if session == nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
            return
        }

        // Set session cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "session",
            Value:    session.ID.String(),
            Path:     "/",
            HttpOnly: true,
            MaxAge:   86400, // 24 hours
        })

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "success",
            "username": foundUser.UserName,
        })
        return

    } else if r.Method == http.MethodGet {
        if err := tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
            fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
            return
        }
    } else {
        fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
        return
    }
}

func Middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := CheckIfLoggedIn(w, r)

		if session == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Extend session
		store.ExtendSession(session.ID)

		// Add session data to the request context
		ctx := context.WithValue(r.Context(), UserSessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Get session ID from cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Parse the UUID from cookie
		if sessionID, err := uuid.FromString(cookie.Value); err == nil {
			store.DeleteSession(sessionID)
		}
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	if r.Method == http.MethodGet {
		err := tmpl.ExecuteTemplate(w, "signup.html", nil)
		if err != nil {
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		name := r.FormValue("username") // Changed from "name" to match the form
		// Create a slice to collect validation errors
		var errors []string

		// Validate email
		if !isValidEmail(email) {
			errors = append(errors, "Invalid email format")
		}

		// Validate username
		if !isValidUsername(name) {
			errors = append(errors, "Username must be 3-30 characters long and contain only letters, numbers, underscores, or hyphens")
		}

		// Validate password (example requirements)
		if len(password) < 8 {
			errors = append(errors, "Password must be at least 8 characters long")
		}

		// If there are any validation errors, return them
		if len(errors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  errors,
			})

			return
		}

		hashedPassword, err := encryptPassword(password)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})

			return
		}

		var user User
		user.Email = email
		user.Password = string(hashedPassword)
		user.UserName = name

		err = SaveUserToDb(user)
		if err != nil {
			// If there's an error (likely user already exists)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
			return
		}

		// Success case
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "success",
			"username": user.UserName,
		})
	}
}

func ReadfromDb() []User {
	users := []User{}
	stmt, err := db.DB.Prepare("SELECT id, username, email, password, created_at FROM users")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return users
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Printf("Error querying users: %v", err)
		return users
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}
	return users
}

func SaveUserToDb(user User) error {
	stmt, err := db.DB.Prepare("INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer stmt.Close()

	user.CreatedAt = time.Now()
	_, err = stmt.Exec(user.UserName, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func CheckIfLoggedIn(w http.ResponseWriter, r *http.Request) *Session {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}

	sessionID, err := uuid.FromString(cookie.Value)
	if err != nil {
		return nil
	}

	session, valid := store.GetSession(sessionID)
	if !valid {
		return nil
	}

	return session
}
