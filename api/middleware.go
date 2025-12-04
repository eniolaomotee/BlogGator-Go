package api

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const userContextkey contextKey = "user"

// AuthMiddleware validates JWT tokens and adds user to context
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get token from header
		token, err := GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error getting bearer token")
			return
		}

		//Validate JWT
		claims, err := ValidateJWT(token, s.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		//Get user from the DB
		userId, err := uuid.Parse(claims.UserId)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid user id in token")
			return
		}

		user, err := s.db.GetUserById(context.Background(), userId)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextkey, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")

		allowOrigin := "*"
		if reqOrigin != "" {
			allowOrigin = reqOrigin
		}

		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		// Also set the non-standard variant in case another layer expects it (debug)
		w.Header().Set("Access-Control-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Logging Middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
