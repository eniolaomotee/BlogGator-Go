package api

import (
	"testing"
	"github.com/google/uuid"
)

func TestMakeandValidateJWT(t *testing.T){
	username := "user"
	userId := uuid.New()
	secret := "secret"

	// create JWT
	token, err := GenerateJWT(userId.String(), username, secret)
	if err != nil{
		t.Fatalf("Error creating JWT: %s", err)
	}

	// validate JWT
	claims, err := ValidateJWT(token,secret)
	if err != nil{
		t.Fatalf("Error validating JWT: %s", err)
	}

	// parse user id
	returnedUserID, err := uuid.Parse(claims.UserId)
	if err != nil{
		t.Fatalf("error parsing user ID %s", err)
	}

	if returnedUserID != userId {
		t.Fatalf("Expected userId %s, got %s", userId, returnedUserID)
	}


}


func TestExpiredJWT(t *testing.T){
	userId := uuid.New()
	username := "user"
	secret := "mysecretkey"


	//Create a token that expires immediately
	token, err := TestGenerateJWT(userId.String(), username, secret)
	if err != nil{
		t.Fatalf("error creating JWT %s", err)
	}

	// validate token has truly expired
	_, err = ValidateJWT(token, secret)
	if err == nil{
		t.Fatalf("Expected error validating expired JWT, got none")
	}
}



func InvalidJWT(t *testing.T){
	secret := "correctsecret"
	wrongSecret := "wrongsecret"
	userId := uuid.New()
	username := "user"


	token, err := GenerateJWT(userId.String(), username,secret)
	if err != nil{
		t.Fatalf("error creating JWT %s", err)
	}

	// validate with wrong secret
	_, err = ValidateJWT(token,wrongSecret)
	if err == nil{
		t.Fatalf("expected error validating JWT with wrong secret, got none")
	}
}