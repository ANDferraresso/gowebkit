package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ANDferraresso/gowebkit/cookie"
)

const CsrfCookieName = "csrf_token"
const CSRFTokenTTL = 30 * time.Minute

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
		Expires:  time.Now().Add(CSRFTokenTTL),
	})
}

func GetCSRFTokenFromCookie(r *http.Request) (string, error) {
	return cookie.Read(r, true, CsrfCookieName)
}

func DeleteCSRFTokenCookie(w http.ResponseWriter) error {
	return cookie.Write(w, true, http.Cookie{
		Name:     CsrfCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
