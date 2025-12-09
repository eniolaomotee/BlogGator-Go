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

	s.router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)

		// Subscription management
		r.Get("/api/billing/subscription", s.handleGetSubscription)
		r.Post("/api/billing/checkout", s.handleCreateCheckout)
		r.Post("/api/billing/cancel", s.handleCancelSusbscription)
		r.Get("/api/billing/tiers", s.handleGetTiers)
		r.Get("/api/billing/usage", s.handleGetUsage)

	})

	// Webhook (no auth - Stripe validates)
	s.router.Post("/api/webhooks/stripe", s.handleStripeWebhook)

	//Protected Routes with limits
	s.router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)
		r.Use(s.TrackAPIUsage)

		// / API access requires Pro+
		r.With(s.CheckAPIAccess).Get("/api/posts", s.handleGetPosts)

		// Feeds
		r.Get("/api/feeds", s.handleGetFeeds)
		r.Post("/api/feeds", s.handleAddFeed)
		r.Post("/api/feeds/follow", s.handleFollowFeed)
		r.Delete("/api/feeds/{feedID}/unfollow", s.handleUnfollowFeed)

		// User Info
		r.Get("/api/me", s.handleGetcurrentUser)

		//Feed creation has limits
		r.With(s.CheckFeedLimit).Post("/api/feeds", s.handleAddFeed)
	})
}

func (s *Server) handleGetTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := s.db.GetAllTiers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching tiers")
		return
	}

	respondWithJson(w, http.StatusOK, tiers)
}

func (s *Server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextkey).(database.User)

	feedCount, _ := s.db.CountUserFeeds(r.Context(), user.ID)
	postCount, _ := s.db.CountUserPosts(r.Context(), user.ID)
	apiCalls, _ := s.db.GetAPIUsageThisMonth(r.Context(), user.ID)

	subscription, err := s.db.GetUserSubscription(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving subscription")
		return
	}

	tier, err := s.db.GetTierByID(r.Context(), subscription.TierID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving tier information")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]interface{}{
		"feeds": map[string]interface{}{
			"current": feedCount,
			"limit":   tier.MaxFeeds,
		},
		"posts": map[string]interface{}{
			"current": postCount,
			"limit":   tier.MaxPosts,
		},
		"api_calls": map[string]interface{}{
			"current": apiCalls,
		},
	})
}
func (s *Server) handleGetcurrentUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextkey).(database.User)

	respondWithJson(w, http.StatusOK, map[string]string{
		"id":         user.ID.String(),
		"username":   user.Name,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})
}
