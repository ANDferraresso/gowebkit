package cookie

const (
	ErrValueTooLong    = "cookie value too long"
	ErrInvalidDuration = "invalid duration"
	ErrInvalidValue    = "invalid cookie value"
	ErrInvSecretKey    = "invalid secret key"
)

/*
import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"encoding/base64"
	"errors"
)


https://go.dev/src/net/http/cookie.go

type Cookie struct {
    Name       string
    Value      string
    Path       string
    Domain     string
    Expires    time.Time
    RawExpires string    // for reading cookies only
    // MaxAge=0 means no 'Max-Age' attribute specified.
    // MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
    // MaxAge>0 means Max-Age attribute present and given in seconds
    MaxAge     int
    Secure     bool
    HttpOnly   bool
    SameSite   http.SameSite
	Raw        string
	Unparsed   []string // Raw text of unparsed attribute-value pairs
}

SameSite
http.SameSiteDefaultMode
http.SameSiteLaxMode
http.SameSiteStrictMode
http.SameSiteNoneMode

// SET COOKIE
cookie := http.Cookie{
    Name:     "exampleCookie",
    Value:    "Hello world!",
    Path:     "/",
    MaxAge:   3600,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
}
http.SetCookie(w, &cookie)

// GET COOKIE
cookie, err := r.Cookie("exampleCookie")
if err != nil {
    switch {
    case errors.Is(err, http.ErrNoCookie):
        http.Error(w, "cookie not found", http.StatusBadRequest)
    default:
        log.Println(err)
        http.Error(w, "server error", http.StatusInternalServerError)
    }
    return
}

*/
