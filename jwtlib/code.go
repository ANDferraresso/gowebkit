package jwtlib

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(secretKey string, hours int64, issuer string, userID string) (string, error) {
	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": issuer,
		"sub": userID,
		"exp": time.Now().Add(time.Duration(hours) * time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(secretKey string, tokenString string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify HMAC (ex. HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	values := map[string]interface{}{
		"iss": "",
		"sub": "",
		"exp": float64(0),
	}

	if err != nil {
		return values, err
	}

	if token.Valid {
		if issuer, ok := claims["iss"].(string); ok {
			values["iss"] = issuer
		}
		if subject, ok := claims["sub"].(string); ok {
			values["sub"] = subject
		}
		if expirationTime, ok := claims["exp"].(float64); ok {
			values["exp"] = expirationTime
		}
	} else {
		return values, fmt.Errorf("invalid JWT")
	}

	return values, nil
}
