package api

import (
	"net/http"
	"time"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) setupRoutes() {
	// Global middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(CORSMiddleware)

	// public routes
	s.router.Post("/api/register", s.handleRegister)
	s.router.Post("/api/login", s.handleLogin)

	//Health check
	s.router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		respondWithJson(w, 200, map[string]string{
			"status": "ok",
		})
	})

	//Protected Routes
	s.router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)

		// Posts
		r.Get("/api/posts", s.handleGetPosts)

		// Feeds
		r.Get("/api/feeds", s.handleGetFeeds)
		r.Post("/api/feeds", s.handleAddFeed)
		r.Post("/api/feeds/follow", s.handleFollowFeed)
		r.Delete("/api/feeds/{feedID}/unfollow", s.handleUnfollowFeed)

		// User Info
		r.Get("/api/me", s.handleGetcurrentUser)
	})
}

func (s *Server) handleGetcurrentUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(database.User)

	respondWithJson(w, http.StatusOK, map[string]string{
		"id":         user.ID.String(),
		"username":   user.Name,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})
}
