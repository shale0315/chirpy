package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testSecret"
	// expires_in := time.Hour

	//Test Creating JWT
	token, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	//Test Validating JWT
	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	//Check if User ID matches
	if validatedID != userID {
		t.Errorf("Expected ID: %v\nActual ID: %v", userID, validatedID)
	}
}

func TestInvalidSecret(t *testing.T) {
	userID := uuid.New()
	secret := "testSecret"
	// expires_in := time.Hour
	wrong_secret := "UndertheFloor!"

	//Test Creating JWT
	token, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	//Test Validating JWT
	_, err = ValidateJWT(token, wrong_secret)
	if err == nil {
		t.Fatalf("Expected to fail with wrong secret, but succeeded")
	}
}

func TestExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "testSecret"
	// expires_in := time.Second

	//Test Creating JWT
	token, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	time.Sleep(2 * time.Second)

	//Test Validating JWT
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("Expected to fail with expired token, but succeeded")
	}

}

func TestBearer(t *testing.T) {
	headers := http.Header{}
	givenToken := "abcde3rqrwq13"
	headers.Set("Authorization", "Bearer "+givenToken)

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	if token != "abcde3rqrwq13" {
		t.Fatalf("incorrect token. expected: %v, but actual: %v", givenToken, token)
	}

}
