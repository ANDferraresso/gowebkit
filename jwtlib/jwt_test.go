package jwtlib

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "test-secret"

func TestGenerateAndVerify(t *testing.T) {
	tok, err := GenerateJWT(secret, 1, "gowebkit", "user-42")
	if err != nil {
		t.Fatal(err)
	}
	v, err := VerifyJWT(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if v["iss"] != "gowebkit" || v["sub"] != "user-42" || v["exp"].(float64) <= float64(time.Now().Unix()) {
		t.Errorf("claims = %v", v)
	}
}

func TestVerifyRejects(t *testing.T) {
	good, _ := GenerateJWT(secret, 1, "i", "s")
	expired, _ := GenerateJWT(secret, -1, "i", "s")

	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "s"}).SignedString([]byte(secret))
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": "s", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "s", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := map[string]struct{ key, token string }{
		"wrong key":    {"other", good},
		"expired":      {secret, expired},
		"no exp":       {secret, noExp},
		"other alg":    {secret, hs512},
		"alg none":     {secret, none},
		"garbage":      {secret, "not.a.token"},
		"empty secret": {"", good},
		"tampered sig": {secret, good[:len(good)-2] + "xx"},
	}
	for name, tt := range tests {
		if _, err := VerifyJWT(tt.key, tt.token); err == nil {
			t.Errorf("%s: token accepted", name)
		}
	}

	if _, err := GenerateJWT("", 1, "i", "s"); err == nil {
		t.Error("empty secret accepted by GenerateJWT")
	}
}
