package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// requestWithCookie returns a request carrying the cookie set by SetCSRFTokenCookie.
func requestWithCookie(t *testing.T, token string, body url.Values) *http.Request {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := SetCSRFTokenCookie(rec, token); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	return req
}

func TestValidateCSRFToken(t *testing.T) {
	token, err := GenerateCSRFToken()
	if err != nil {
		t.Fatal(err)
	}

	req := requestWithCookie(t, token, url.Values{CSRFFormField: {token}})
	if !ValidateRequest(req) {
		t.Error("valid form token rejected")
	}

	req = requestWithCookie(t, token, url.Values{})
	req.Header.Set(CSRFHeaderName, token)
	if !ValidateRequest(req) {
		t.Error("valid header token rejected")
	}

	req = requestWithCookie(t, token, url.Values{CSRFFormField: {"wrong"}})
	if ValidateRequest(req) {
		t.Error("wrong token accepted")
	}

	req = requestWithCookie(t, token, url.Values{})
	if ValidateRequest(req) {
		t.Error("missing submitted token accepted")
	}

	req = httptest.NewRequest("POST", "/", strings.NewReader(url.Values{CSRFFormField: {token}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if ValidateRequest(req) {
		t.Error("request without cookie accepted")
	}

	req = requestWithCookie(t, "", url.Values{CSRFFormField: {""}})
	if ValidateRequest(req) {
		t.Error("empty tokens accepted")
	}
}
