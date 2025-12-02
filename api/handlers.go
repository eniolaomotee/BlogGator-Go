package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	db *database.Queries
	router *chi.Mux
}

func NewServer(db *database.Queries) *Server{
	s := &Server{
		db: db,
		router: chi.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request){
	s.router.ServeHTTP(w,r)
}

// Handlers
func (s *Server) handleRegister (w http.ResponseWriter, r *http.Request){
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		respondWithError(w,http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondWithError(w,http.StatusUnauthorized, "Username and password are required")
		return
	}

	//Hash Password
	passwordHash, err := Hashpassword(req.Password)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Error hashing Password")
		return
	}

	// Create user
	user, err := s.db.CreateUserWithPassword(context.Background(),database.CreateUserWithPasswordParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: req.Username,
		PasswordHash: passwordHash,
	})
	if err != nil{
		if strings.Contains(err.Error(), "duplicate key"){
			respondWithError(w, http.StatusConflict, "Username already exists")
			return 
		}
		respondWithError(w, http.StatusInternalServerError, "error creating user")
		return
	}

	// Generate JWT
	userJwt, err := GenerateJWT(user.ID.String(), user.Name)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "error generating token")
	}

	respondWithJson(w, http.StatusCreated, AuthResponse{
		Token: userJwt,
		UserID: user.ID.String(),
		Username: user.Name,
	})
}



// Handle login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request){
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		respondWithError(w,http.StatusBadRequest, "Invalid Request body")
		return
	}

	// Get user
	user, err := s.db.GetUser(context.Background(), req.Username)
	if err != nil{
		if err == sql.ErrNoRows{
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		}
		respondWithError(w, http.StatusInternalServerError, "error getting user")
		return
	}

	// check password
	if !CheckPasswordWithHash(req.Password, user.PasswordHash){
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	//generate JWT
	token, err := GenerateJWT(user.ID.String(), user.Name)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "error generating token")
		return
	}

	respondWithJson(w,http.StatusOK, AuthResponse{
		Token: token,
		UserID: user.ID.String(),
		Username: user.Name,
	})

}



// Handler for get posts
func (s *Server) handleGetPosts(w http.ResponseWriter, r *http.Request){
	// get user from context (set by auth middleware)
	user := r.Context().Value("user").(database.User)


	// parse query params
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != ""{
		fmt.Sscanf(limitStr, "%d",&limit)
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == ""{
		sortBy = "published_at"
	}

	orderBy := r.URL.Query().Get("order")
	if orderBy == ""{
		orderBy = "desc"
	}

	feedFilter := r.URL.Query().Get("feed")


	//fetch posts
	posts, err := s.db.GetPostsForUserSorted(context.Background(), database.GetPostsForUserSortedParams{
		UserID: user.ID,
		Limit: int32(limit),
		Column3: feedFilter,
		Column4: sortBy + "_" + orderBy,
		Offset: 0,
	})
	if err != nil{
		respondWithError(w,http.StatusInternalServerError, "Error fetching posts")
		return
	}

	//Convert to response format
	response := make([]PostResponse, len(posts))
	for i, post := range posts{
		var desc *string
		if post.Description.Valid {
			desc = &post.Description.String
		}

		response[i] = PostResponse{
			ID: post.ID.String(),
			Title: post.Title,
			Url: post.Url,
			Description: desc,
			PublishedAt: post.PublishedAt.Time.Format(time.RFC3339),
			FeedName: post.FeedName,
		}
	}

	respondWithJson(w, http.StatusOK, response)

}

// Handle get feeds
func (s *Server) handleGetFeeds(w http.ResponseWriter, r *http.Request){
	// get user from context (set by auth middleware)
	user := r.Context().Value("user").(database.User)

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError,"error fetching feeds")
		return
	}

	response := make([]FeedResponse, len(feeds))
	for i, feed := range feeds{
		response[i] = FeedResponse{
			ID: feed.ID.String(),
			Name: feed.FeedName,
			CreatedAt: feed.CreatedAt.Format(time.RFC3339),
		}
	}

	respondWithJson(w, http.StatusOK, response)

}


// Handle Addfeed
func (s *Server) handleAddFeed(w http.ResponseWriter, r *http.Request){
	user := r.Context().Value("user").(database.User)

	var req struct{
		Name string `json:"name"`
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		respondWithError(w,http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.URL == ""{
		respondWithError(w, http.StatusUnauthorized, "Name and URL are required")
		return
	}


	// Create feed to add to DB
	feed, err := s.db.CreateFeed(context.Background(),database.CreateFeedParams{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: req.Name,
		Url: req.URL,
		UserID: user.ID,
	})
	if err != nil{
		if strings.Contains(err.Error(), "duplicate key"){
			respondWithError(w,http.StatusConflict, "Feed already exists")
		}
		respondWithError(w, http.StatusInternalServerError, "unable to create feed")
		return
	}

	// Auto-follow feed
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Error following feed")
		return
	}

	respondWithJson(w,http.StatusOK, FeedResponse{
		ID: feed.ID.String(),
		Name: feed.Name,
		URL: feed.Url,
		CreatedAt: feed.CreatedAt.Format(time.RFC3339),
	})

}


// Handle FollowFeed
func (s *Server) handleFollowFeed(w http.ResponseWriter, r *http.Request){
	user := r.Context().Value("user").(database.User)

	var req struct{
		FeedId string `json:"feed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	feedId, err := uuid.Parse(req.FeedId)
	if err != nil{
		respondWithJson(w, http.StatusBadRequest, "Invalid feed ID")
		return
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedId,
	})
	if err != nil{
		if strings.Contains(err.Error(),"duplicate key"){
			respondWithError(w, http.StatusConflict, "Already following this feed")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error following feed")
		return
	}

	respondWithJson(w,http.StatusCreated, map[string]string{
		"message":"successfully followed feed",
	})
}

// Handle Unfollowfeed
func (s *Server) handleUnfollowFeed(w http.ResponseWriter, r *http.Request){
	user := r.Context().Value("user").(database.User)

	feedIdstr := chi.URLParam(r, "feedID")
	feedId, err := uuid.Parse(feedIdstr)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "invalid feed id")
		return
	}

	err = s.db.DeleteFeedFollowByUserAndFeed(context.Background(),database.DeleteFeedFollowByUserAndFeedParams{
		UserID: user.ID,
		FeedID: feedId,
	})
	if err != nil{
		respondWithError(w,http.StatusInternalServerError, "Error unfollowing feed")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{
		"message":"Successfully Unfollowed feed",
	})

}