package jwtlib

import (
	"testing"
	"time"
)

func TestGenerateAndVerifyJWT(t *testing.T) {
	secret := "supersecretkey"
	userID := "12345"

	// Generazione del token
	token, err := GenerateJWT(secret, userID)
	if err != nil {
		t.Fatalf("Errore nella generazione del JWT: %v", err)
	}
	if token == "" {
		t.Fatal("Token generato vuoto")
	}

	// Verifica del token
	claims, err := VerifyJWT(secret, token)
	if err != nil {
		t.Fatalf("Errore nella verifica del JWT: %v", err)
	}

	// Controllo dei claim
	if claims["iss"] != "NewDida" {
		t.Errorf("Claim 'iss' errato: got %v, want %v", claims["iss"], "NewDida")
	}

	if claims["sub"] != userID {
		t.Errorf("Claim 'sub' errato: got %v, want %v", claims["sub"], userID)
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Errorf("Claim 'exp' non è float64: %v", claims["exp"])
	} else if int64(expFloat) < time.Now().Unix() {
		t.Errorf("Token già scaduto: exp = %v", expFloat)
	}
}
