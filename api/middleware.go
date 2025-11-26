package api

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens and adds user to context
func (s *Server) AuthMiddleware (next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == ""{
			respondWithError(w, http.StatusUnauthorized,"Missing authorization header")
			return
		}

		// Bearer token check 
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer"{
			respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		//Validate JWT
		claims, err := ValidateJWT(token)
		if err != nil{
			respondWithError(w,http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		//Get user from the DB
		userId, err := uuid.Parse(claims.UserId)
		if err != nil{
			respondWithError(w, http.StatusUnauthorized, "Invalid user id in token")
			return 
		}

		user, err := s.db.GetUserById(context.Background(),userId)
		if err != nil{
			respondWithError(w,http.StatusUnauthorized, "User not found")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w,r.WithContext(ctx))

	})
}


// CORS middleware 
func CORSMiddleware(next http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT, DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,Authorization")

		if r.Method == "OPTIONS"{
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w,r)
	})
}

// Logging Middleware
func LoggingMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s",r.Method,r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w,r)
	})
}