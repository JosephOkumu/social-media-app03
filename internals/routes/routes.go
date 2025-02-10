package routes

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"forum/internals/auth"
	"forum/internals/comments"
	"forum/internals/fails"
	"forum/internals/post"
)

func RegisteringRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", post.ServeHomePage)
	mux.HandleFunc("/about", post.ServeAboutPage)

	// Post Routes.
	mux.HandleFunc("/posts", post.ServePosts)
	mux.HandleFunc("/view-post", post.ViewPost)
	mux.HandleFunc("/create-post-form", auth.Middleware(http.HandlerFunc(post.ServeCreatePostForm)))
	mux.HandleFunc("/upload-image", auth.Middleware(http.HandlerFunc(post.UploadImage)))
	mux.HandleFunc("/categories", post.ServeCategories)
	mux.HandleFunc("/create-post", auth.Middleware(http.HandlerFunc(post.CreatePost)))
	mux.HandleFunc("/post/react", auth.Middleware(http.HandlerFunc(post.ReactToPost)))

	// Auth Routes.
	mux.HandleFunc("/signup", auth.Signup)
	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)

	// Google Auth Routes.
	mux.HandleFunc("/auth/google", auth.InitiateGoogleAuth)
	mux.HandleFunc("/auth/google/callback", auth.HandleGoogleCallback)

	// GitHub Auth Routes.
	mux.HandleFunc("/auth/github", auth.InitiateGitHubAuth)
	mux.HandleFunc("/auth/github/callback", auth.HandleGitHubCallback)

	// Facebook Auth Routes
	mux.HandleFunc("/auth/facebook/login", auth.InitiateFacebookAuth)
	mux.HandleFunc("/auth/facebook/callback", auth.HandleFacebookCallback)

	// Filter Routes.
	mux.HandleFunc("/category", post.ViewPostsByCategory)
	mux.HandleFunc("/userfilter", auth.Middleware(http.HandlerFunc(post.FilterbyUser)))
	mux.HandleFunc("/likesfilter", auth.Middleware(http.HandlerFunc(post.FilterbyLikes)))

	// Comment Routes
	mux.HandleFunc("/comments", comments.GetComments)
	mux.HandleFunc("/comments/create", auth.Middleware(http.HandlerFunc(comments.CreateComment)))
	mux.HandleFunc("/comments/react", auth.Middleware(http.HandlerFunc(comments.ReactToComment)))

	// static
	mux.HandleFunc("/static/", serveStatic)
	return mux
}

// serveStatic serves static files
func serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}
	// Remove the /static/ prefix from the URL path
	filePath := path.Join("static", strings.TrimPrefix(r.URL.Path, "/static/"))

	// Check if the file exists and is not a directory
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		fails.ErrorPageHandler(w, r, http.StatusForbidden)
		return
	}

	// Check the file extension and set the appropriate Content-Type
	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".bmp":
		w.Header().Set("Content-Type", "image/bmp")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		fails.ErrorPageHandler(w, r, http.StatusForbidden)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}
