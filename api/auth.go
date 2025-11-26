package api

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret string

func init() {
	godotenv.Load()
	jwtSecret = os.Getenv("SECRET_KEY")
}

type Claims struct{
	UserId string `json:"user_id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

//Hashpassword creates a bcrypt hash of the password
func Hashpassword(password string) (string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
// Check password compared with Hash
func CheckPasswordWithHash(password, hash string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err == nil
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userId, username string) (string, error){
	claims := Claims{
		UserId: userId,
		UserName: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt : jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return token.SignedString([]byte(jwtSecret))
}


// ValidateJWT validates a JWT token and return the claims
func ValidateJWT(tokenString string) (*Claims, error){
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil{
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid{
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}


// GenerateAPIKey generates a random API key
func GenerateAPIKey()(string,error){
	bytes := make([]byte,32)
	if _, err := rand.Read(bytes); err != nil{
		return "", err
	}
	return  hex.EncodeToString(bytes), nil
}