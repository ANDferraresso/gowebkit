package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ANDferraresso/gowebkit/cookie"
)

const CsrfCookieName = "csrf_token"
const CSRFTokenTTL = 30 * time.Minute

// CSRFFormField e CSRFHeaderName sono i punti in cui TokenFromRequest
// cerca il token inviato dal client.
const (
	CSRFFormField  = "_csrf"
	CSRFHeaderName = "X-CSRF-Token"
)

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

// TokenFromRequest restituisce il token inviato dal client: prima l'header
// X-CSRF-Token (per le richieste AJAX), poi il campo di form "_csrf".
func TokenFromRequest(r *http.Request) string {
	if t := r.Header.Get(CSRFHeaderName); t != "" {
		return t
	}
	return r.PostFormValue(CSRFFormField)
}

// ValidateCSRFToken confronta il token inviato (submitted) con quello salvato
// nel cookie, con un confronto a tempo costante. Ritorna false se uno dei due
// manca. Schema "double submit cookie": il token va generato con
// GenerateCSRFToken, salvato con SetCSRFTokenCookie e inserito nel form
// (es. form.AddCsrfField).
func ValidateCSRFToken(r *http.Request, submitted string) bool {
	expected, err := GetCSRFTokenFromCookie(r)
	if err != nil || expected == "" || submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(submitted)) == 1
}

// ValidateRequest è una scorciatoia per ValidateCSRFToken(r, TokenFromRequest(r)).
func ValidateRequest(r *http.Request) bool {
	return ValidateCSRFToken(r, TokenFromRequest(r))
}
