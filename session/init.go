package session

import (
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"time"
	// "github.com/google/uuid"
)

const (
	ErrSessionExpired        = "session expired"
	ErrSessionAlreadyExists  = "session already exists"
	ErrSessionNotFound       = "session not found"
	ErrKeyNotFound           = "key not found"
	ErrWhileConnectingMysql  = "unable to connect to MySQL"
	ErrUnableSerializeJson   = "unable to serialize in json format"
	ErrSqlInsertQuery        = "error in SQL insert query"
	ErrSqlSelectQuery        = "error in SQL select query"
	ErrSqlUpdateQuery        = "error in SQL update query"
	ErrSqlDeleteQuery        = "error in SQL delete query"
	ErrParsingExpiresAt      = "error while parsing expires_at"
	ErrUnableUnserializeJson = "unable to unserialize json data"
	ErrInvTableName          = "invalid table name"
)

const dateFormat = "2006-01-02 15:04:05"

var validTableName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type SessionData map[string]interface{}

/*

type Session struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	ExpiresAt time.Time              `json:"expiresAt"`
}
*/

type Store interface {
	CreateStore() string
	InitStore(interface{}, string) error
	Init(time.Time) (string, error)
	Check(string) bool
	Get(sid string, key string) (interface{}, error)
	Set(sid string, key string, value interface{}) error
	RemoveKey(sid string, key string) error
	Delete(sid string) error
	Refresh(sid string) (string, error)
}

func generateSessionID() (string, error) {
	// uuid.NewString()
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// return uuid.NewString()
	return base64.URLEncoding.EncodeToString(b), nil
}
