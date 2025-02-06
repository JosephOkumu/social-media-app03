package routes

import (
	"net/http"

	"forum/internals/auth"
	"forum/internals/comments"
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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}
