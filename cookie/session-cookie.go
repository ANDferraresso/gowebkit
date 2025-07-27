package cookie

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultSessionCookieName = "SESSION_ID"
)

func WriteSessionCookie(w http.ResponseWriter, sKey string, value string, duration string) error {
	var h int
	h, err := strconv.Atoi(duration)
	if err != nil {
		return errors.New(ErrInvalidDuration)
	}

	cookie := http.Cookie{
		Name:   DefaultSessionCookieName,
		Value:  value,
		Path:   "/", // string
		Domain: "",  // string
		// Expires time.Time
		// MaxAge=0 means no 'Max-Age' attribute specified.
		// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
		// MaxAge>0 means Max-Age attribute present and given in seconds
		// Expires: Usa una data e un orario assoluti (formato RFC1123) per indicare la scadenza del cookie.
		// MaxAge: Usa un intervallo in secondi che specifica per quanto tempo il cookie è valido a partire da ora
		// MaxAge: 3 * 60 * 60, // Scadenza tra 3 ore in secondi
		// Expires: time.Now().Add(3 * time.Hour), // Scadenza tra 3 ore
		// Se entrambi i campi sono impostati, MaxAge ha la precedenza su Expires.
		Expires:  time.Now().Add(time.Duration(h) * time.Hour), // Scadenza tra 3 ore
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // http.SameSite
		/*
			Gli sviluppatori devono utilizzare una nuova impostazione dei cookie,
			SameSite=None, per contrassegnare i cookie per l'accesso cross-site;
			quando è presente l'attributo SameSite=None, è necessario utilizzare
			un attributo Secure aggiuntivo in modo che i cookie cross-site
			siano accessibili solo tramite connessioni HTTPS.
		*/
	}

	secretKey, err := hex.DecodeString(sKey)
	if err != nil {
		return errors.New(ErrInvSecretKey)
	}

	return WriteSigned(w, cookie, secretKey)
	// http.SetCookie(w, &cookie)
}

func ReadSessionCookie(r *http.Request, sKey string) (string, error) {
	secretKey, err := hex.DecodeString(sKey)
	if err != nil {
		return "", errors.New(ErrInvSecretKey)
	}

	return ReadSigned(r, DefaultSessionCookieName, secretKey)
}
