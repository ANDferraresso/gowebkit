package cookie

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var key32 = bytes.Repeat([]byte{7}, 32)

// roundTrip writes a cookie with write, then returns a request carrying it
// (optionally tampered by mutate).
func roundTrip(t *testing.T, write func(w http.ResponseWriter) error, mutate func(*http.Cookie)) *http.Request {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := write(rec); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range rec.Result().Cookies() {
		if mutate != nil {
			mutate(c)
		}
		req.AddCookie(c)
	}
	return req
}

func TestPlainCookie(t *testing.T) {
	req := roundTrip(t, func(w http.ResponseWriter) error {
		return Write(w, true, http.Cookie{Name: "c", Value: "hello; world"})
	}, nil)
	if v, err := Read(req, true, "c"); err != nil || v != "hello; world" {
		t.Errorf("Read = %q, %v", v, err)
	}

	rec := httptest.NewRecorder()
	if err := Write(rec, false, http.Cookie{Name: "c", Value: strings.Repeat("a", 5000)}); err == nil {
		t.Error("cookie over 4096 bytes accepted")
	}
}

func TestSignedCookie(t *testing.T) {
	write := func(w http.ResponseWriter) error {
		return WriteSigned(w, http.Cookie{Name: "s", Value: "user=42"}, key32)
	}
	if v, err := ReadSigned(roundTrip(t, write, nil), "s", key32); err != nil || v != "user=42" {
		t.Errorf("ReadSigned = %q, %v", v, err)
	}
	if _, err := ReadSigned(roundTrip(t, write, nil), "s", bytes.Repeat([]byte{8}, 32)); err == nil {
		t.Error("wrong key accepted")
	}
	tamper := func(c *http.Cookie) { c.Value = c.Value[:len(c.Value)-4] + "AAAA" }
	if _, err := ReadSigned(roundTrip(t, write, tamper), "s", key32); err == nil {
		t.Error("tampered cookie accepted")
	}
	short := func(c *http.Cookie) { c.Value = "YWJj" } // "abc"
	if _, err := ReadSigned(roundTrip(t, write, short), "s", key32); err == nil {
		t.Error("short cookie accepted")
	}
}

func TestEncryptedCookie(t *testing.T) {
	write := func(w http.ResponseWriter) error {
		return WriteEncrypted(w, http.Cookie{Name: "e", Value: "secret:data"}, key32)
	}
	if v, err := ReadEncrypted(roundTrip(t, write, nil), "e", key32); err != nil || v != "secret:data" {
		t.Errorf("ReadEncrypted = %q, %v", v, err)
	}
	if _, err := ReadEncrypted(roundTrip(t, write, nil), "e", bytes.Repeat([]byte{8}, 32)); err == nil {
		t.Error("wrong key accepted")
	}
	// A cookie encrypted for name "e" must not be accepted under another name.
	rename := func(c *http.Cookie) { c.Name = "other" }
	if _, err := ReadEncrypted(roundTrip(t, write, rename), "other", key32); err == nil {
		t.Error("cookie accepted under a different name")
	}
}

func TestSessionCookie(t *testing.T) {
	sKey := hex.EncodeToString(key32)
	write := func(w http.ResponseWriter) error { return WriteSessionCookie(w, sKey, "sid-123", 3) }
	if v, err := ReadSessionCookie(roundTrip(t, write, nil), sKey); err != nil || v != "sid-123" {
		t.Errorf("ReadSessionCookie = %q, %v", v, err)
	}
	if err := WriteSessionCookie(httptest.NewRecorder(), "not-hex", "x", 1); err == nil {
		t.Error("invalid hex key accepted")
	}
}
