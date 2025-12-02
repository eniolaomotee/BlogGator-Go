package api

// Request/ Response Types
type RegisterRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct{
	Username string `json:"username"`
	Password string `json:"password"` 
}

type AuthResponse struct{
	Token string `json:"token"`
	UserID string `json:"user_id"`
	Username string `json:"username"`
}

type ErrorResponse struct{
	Error string `json:"error"`
}

type PostResponse struct {
	ID string  `json:"id"`
	Title string `json:"title"`
	Url string `json:"url"`
	Description *string `json:"description"`
	PublishedAt string `json:"published_at"`
	FeedName string `json:"feed_name"`
}

type FeedResponse struct{
	ID  string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type Request struct {
	Name string `json:"name"`
	URL string `json:"url"`
}