package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/ANDferraresso/gowebkit/cookie"
)

const CsrfCookieName = "csrf_token"

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func SetCSRFTokenCookie(w http.ResponseWriter, token string) error {
	return cookie.Write(w, true, http.Cookie{
		Name:     CsrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetCSRFTokenFromCookie(r *http.Request) (string, error) {
	return cookie.Read(r, true, CsrfCookieName)
}
