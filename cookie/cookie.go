package cookie

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// https://www.alexedwards.net/blog/working-with-cookies-in-go

func Write(w http.ResponseWriter, base64Enc bool, cookie http.Cookie) error {
	if base64Enc {
		cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))
	}

	if len(cookie.String()) > 4096 {
		return errors.New(ErrValueTooLong)
	}

	http.SetCookie(w, &cookie)
	return nil
}

func Read(r *http.Request, base64Dec bool, name string) (string, error) {
	// Read the cookie as normal.
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	if base64Dec {
		value, err := base64.URLEncoding.DecodeString(cookie.Value)
		if err != nil {
			return "", errors.New(ErrInvalidValue)
		}
		return string(value), nil
	}

	return cookie.Value, nil
}

func DeleteCookie(w http.ResponseWriter, name string, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path, // "/" se vuoi cancellare su tutto il dominio
		MaxAge:   -1,   // forza la scadenza immediata
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // o quello che usi
	})
}

// Esempio:
// import "encoding/hex"
// secretKey, err = hex.DecodeString("13d6b4dff8f84a10851021ec8608f814570d562c92fe6b5ec4c9f595bcb3234b")
// 64-character hex string to give us a slice containing 32 random bytes
func WriteSigned(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	// Calculate a HMAC signature of the cookie name and value, using SHA256 and
	// a secret key (which we will create in a moment).
	// cookie.Value = "{HMAC signature}{original value}"
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend the cookie value with the HMAC signature.
	// cookie.Value = string(signature) + cookie.Value // signature is a sequence of arbitrary bytes, and converting it with string() may introduce invalid or ambiguous characters into the cookie context.
	sigEncoded := base64.URLEncoding.EncodeToString(signature)
	cookie.Value = sigEncoded + cookie.Value

	// Call our Write() helper to base64-encode the new cookie value and write
	// the cookie.
	return Write(w, true, cookie)
}

func ReadSigned(r *http.Request, name string, secretKey []byte) (string, error) {
	// Read in the signed value from the cookie. This should be in the format
	// "{HMAC signature}{original value}".
	signedValue, err := Read(r, true, name)
	if err != nil {
		return "", err
	}

	// A SHA256 HMAC signature has a fixed length of 32 bytes. To avoid a potential
	// 'index out of range' panic in the next step, we need to check sure that the
	// length of the signed cookie value is at least this long. We'll use the
	// sha256.Size constant here, rather than 32, just because it makes our code
	// a bit more understandable at a glance.
	if len(signedValue) < sha256.Size {
		return "", errors.New(ErrInvalidValue)
	}

	// Split apart the signature and original cookie value.
	// signature := signedValue[:sha256.Size]
	// value := signedValue[sha256.Size:]
	sigLen := base64.URLEncoding.EncodedLen(sha256.Size)
	if len(signedValue) < sigLen {
		return "", errors.New(ErrInvalidValue)
	}

	sigEncoded := signedValue[:sigLen]
	value := signedValue[sigLen:]
	signature, err := base64.URLEncoding.DecodeString(sigEncoded)
	if err != nil {
		return "", errors.New(ErrInvalidValue)
	}

	// Recalculate the HMAC signature of the cookie name and original value.
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Check that the recalculated signature matches the signature we received
	// in the cookie. If they match, we can be confident that the cookie name
	// and value haven't been edited by the client.
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", errors.New(ErrInvalidValue)
	}

	// Return the original cookie value.
	return value, nil
}

func WriteEncrypted(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create a unique nonce containing 12 random bytes.
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}

	// Prepare the plaintext input for encryption. Because we want to
	// authenticate the cookie name as well as the value, we make this plaintext
	// in the format "{cookie name}:{cookie value}". We use the : character as a
	// separator because it is an invalid character for cookie names and
	// therefore shouldn't appear in them.
	plaintext := fmt.Sprintf("%s:%s", cookie.Name, cookie.Value)

	// Encrypt the data using aesGCM.Seal(). By passing the nonce as the first
	// parameter, the encrypted data will be appended to the nonce — meaning
	// that the returned encryptedValue variable will be in the format
	// "{nonce}{encrypted plaintext data}".
	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Set the cookie value to the encryptedValue.
	// cookie.Value = string(encryptedValue)
	cookie.Value = base64.URLEncoding.EncodeToString(encryptedValue)

	// Write the cookie as normal.
	return Write(w, true, cookie)
}

func ReadEncrypted(r *http.Request, name string, secretKey []byte) (string, error) {
	// Read the encrypted value from the cookie as normal.
	// encryptedValue, err := Read(r, true, name)
	encryptedValueStr, err := Read(r, true, name)
	if err != nil {
		return "", err
	}
	encryptedValue, err := base64.URLEncoding.DecodeString(encryptedValueStr)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Get the nonce size.
	nonceSize := aesGCM.NonceSize()

	// To avoid a potential 'index out of range' panic in the next step, we
	// check that the length of the encrypted value is at least the nonce
	// size.
	if len(encryptedValue) < nonceSize {
		return "", errors.New(ErrInvalidValue)
	}

	// Split apart the nonce from the actual encrypted data.
	nonce := encryptedValue[:nonceSize]
	ciphertext := encryptedValue[nonceSize:]

	// Use aesGCM.Open() to decrypt and authenticate the data. If this fails,
	// return a ErrInvalidValue error.
	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", errors.New(ErrInvalidValue)
	}

	// The plaintext value is in the format "{cookie name}:{cookie value}". We
	// use strings.Cut() to split it on the first ":" character.
	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		return "", errors.New(ErrInvalidValue)
	}

	// Check that the cookie name is the expected one and hasn't been changed.
	if expectedName != name {
		return "", errors.New(ErrInvalidValue)
	}

	// Return the plaintext cookie value.
	return value, nil
}
