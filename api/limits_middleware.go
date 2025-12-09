package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
)

func (s *Server) CheckFeedLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextkey).(database.User)

		// Get user's subscription
		subscription, err := s.db.GetUserSubscription(context.Background(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error retrieving subscription")
			return
		}

		// Get tier limits
		tier, err := s.db.GetTierByID(context.Background(), subscription.TierID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error checking limits")
			return
		}

		// count user's current feeds
		feedCount, err := s.db.CountUserFeeds(context.Background(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error counting user's feeds")
			return
		}

		//Check limit
		if feedCount >= int64(tier.MaxFeeds) {
			respondWithError(w, http.StatusForbidden, "Feed limit reached. Upgrade for more feeds")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) CheckAPIAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextkey).(database.User)

		subscription, err := s.db.GetUserSubscription(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusForbidden,
				"API access requires a Pro subscription or higher")
			return
		}

		tier, err := s.db.GetTierByID(r.Context(), subscription.TierID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error checking access")
			return
		}

		// Check if tier has API access
		var features map[string]interface{}
		err = json.Unmarshal(tier.Features, &features)
		if err != nil || features["api_access"] != true {
			respondWithError(w, http.StatusForbidden,
				"API access not available on your plan. Upgrade to Pro.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) TrackAPIUsage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextkey).(database.User)

		// Track API call
		go func() {
			_ = s.db.IncrementAPIUsage(context.Background(), user.ID)
		}()

		next.ServeHTTP(w, r)
	})
}
